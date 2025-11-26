// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	chanSendTimeout = 100 * time.Millisecond

	// time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second
	// send pings to peer with this period. Must be less than pongWait.
	pingPeriod = pongWait * 9 / 10

	wsScheme    = "ws"
	wsSchemeTLS = "wss"

	httpScheme    = "http"
	httpSchemeTLS = "https"
)

var (
	SubscriptionTimeout = 3 * time.Second
)

// Errors
var (
	ErrBadWrapperObject = errors.New("invalid wrapper object")
	ErrTimeout          = errors.New("timeout")
)

// wrapperObject is an internal abstraction of the websockets data exchange object.
type wrapperObject struct {
	// Embedded error fields. Will only be present on error: `event` is error.
	WebsocketError

	// Wrapper object of websockets data exchanged between client and server
	WrapperObject
}

var _ slog.LogValuer = (*wrapperObject)(nil)

// Connect will attempt to connect to the Omlox™ Hub websockets interface.
func Connect(ctx context.Context, addr string, options ...ClientOption) (*Client, error) {
	c, err := New(addr, options...)
	if err != nil {
		return nil, err
	}

	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

// Connect dials the Omlox™ Hub websockets interface.
func (c *Client) Connect(ctx context.Context) error {
	if !c.isClosed() {
		// close the connection if it happens to be open
		if err := c.Close(); err != nil {
			return err
		}
	}

	wsURL := c.baseAddress.JoinPath("/ws/socket")

	if err := upgradeToWebsocketScheme(wsURL); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)

	conn, _, err := websocket.Dial(ctx, wsURL.String(), &websocket.DialOptions{
		HTTPClient: c.client,
	})
	if err != nil {
		cancel()
		return err
	}

	slog.LogAttrs(
		ctx,
		slog.LevelDebug,
		"connected",
		slog.String("host", wsURL.Hostname()),
		slog.String("path", wsURL.EscapedPath()),
		slog.Bool("secured", wsURL.Scheme == wsSchemeTLS),
	)

	c.mu.Lock()
	c.conn = conn
	c.closed = false
	c.cancel = cancel
	c.lifecycleWg.Add(2)
	c.mu.Unlock()

	go func() {
		defer c.lifecycleWg.Done()
		if err := c.readLoop(ctx); err != nil {
			return
		}
	}()

	go func() {
		defer c.lifecycleWg.Done()
		if err := c.pingLoop(ctx); err != nil {
			return
		}
	}()

	return nil
}

// Publish a message to the Omlox Hub.
func (c *Client) Publish(ctx context.Context, topic Topic, payload ...json.RawMessage) error {
	if topic == "" {
		return errors.New("empty topic")
	}

	wrObj := &WrapperObject{
		Event:   EventMsg,
		Topic:   topic,
		Payload: payload,
	}

	return c.publish(ctx, wrObj)
}

func (c *Client) publish(ctx context.Context, wrObj *WrapperObject) (err error) {
	// TODO @dvcorreia: maybe this log should be a metric instead.
	defer slog.LogAttrs(context.Background(), slog.LevelDebug, "published", slog.Any("err", err), slog.Any("event", wrObj))

	if c.isClosed() {
		return net.ErrClosed
	}

	// TODO @dvcorreia: use the easyjson marshal method.
	return wsjson.Write(ctx, c.conn, wrObj)
}

// Subscribe to a topic in Omlox Hub.
func (c *Client) Subscribe(ctx context.Context, topic Topic, params ...Parameter) (*Subcription, error) {
	parameters := make(Parameters)
	for _, param := range params {
		if err := param(topic, parameters); err != nil {
			return nil, err
		}
	}

	return c.subscribe(ctx, topic, parameters)
}

// Sends a subscription message and handles the confirmation from the server.
//
// The subscription will be attributed an ID that can used for futher context.
// There can only be one pending subscription at each time.
// Subsequent subscriptions will wait while the pending one is waiting for an ID from the server.
// Since each subscription on a topic can have a distinct parameters, we must synchronisly wait to match each one to its ID.
func (c *Client) subscribe(ctx context.Context, topic Topic, params Parameters) (*Subcription, error) {
	// channel to await subscription confirmation
	await := make(chan struct {
		sid int
		err error
	})
	defer close(await)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	// lock for pending subscription confirmation.
	// the pending will be freed by the subribed message handler.
	case c.pending <- await:
	}

	wrObj := &WrapperObject{
		Event:   EventSubscribe,
		Topic:   topic,
		Params:  params,
		Payload: nil,
	}

	if err := c.publish(ctx, wrObj); err != nil {
		<-c.pending // clear pending subscription
		return nil, err
	}

	// wait for subcription ID
	var r struct {
		sid int
		err error
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r = <-await:
	}

	if r.err != nil {
		return nil, r.err
	}

	sub := &Subcription{
		// sid:    r.sid,
		sid:    0, // BUG: deephub doesn't return the sid in subsequent messages (NEEDS FIX!)
		topic:  topic,
		params: params,
		mch:    make(chan *WrapperObject, 1),
	}

	// promote a pending subcription
	c.mu.Lock()
	c.subs[sub.sid] = sub
	c.mu.Unlock()

	return sub, nil
}

// ping pong loop that manages the websocket connection health.
func (c *Client) pingLoop(ctx context.Context) error {
	t := time.NewTicker(pingPeriod)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}

		ctx, cancel := context.WithTimeout(ctx, pongWait)
		defer cancel()

		begin := time.Now()
		err := c.conn.Ping(ctx)

		if err != nil {
			// context was exceded and the client should close
			if errors.Is(err, context.Canceled) {
				return nil
			}

			if errors.Is(err, context.DeadlineExceeded) { // TODO @dvcorreia: redundant?
				// ping could not be done, context exceded and connecting will be closed
				// reconnect or close the client
				return err
			}

			return err
		}

		slog.Debug("heartbeat", slog.Duration("latency", time.Since(begin)))
	}
}

// readLoop that will handle incomming data.
func (c *Client) readLoop(ctx context.Context) error {
	defer c.clearSubs()

	// set the client to closed state
	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.closed = true
	}()

	for {
		msgType, r, err := c.conn.Reader(ctx)

		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}

			// when received StatusNormalClosure or StatusGoingAway close frame will be translated to io.EOF when reading
			// the TCP connection can also return it, which is passed down from the lib to here
			if errors.Is(err, io.EOF) {
				return nil
			}

			// when the connection is closed, due to something (e.g. ping deadline, etc)
			var e net.Error
			if errors.As(err, &e) {
				return nil
			}

			switch s := websocket.CloseStatus(err); s {
			case websocket.StatusGoingAway, websocket.StatusNormalClosure:
				return nil
			}

			return err
		}

		// assumed that messages can only betext until further specification.
		if msgType != websocket.MessageText {
			continue
		}

		var (
			wrObj wrapperObject
			d     = json.NewDecoder(r) // TODO @dvcorreia: maybe use easyjson
		)
		if err := d.Decode(&wrObj); err != nil {
			// TODO @dvcorreia: print debug logs or provide metrics
			continue
		}

		slog.LogAttrs(context.Background(), slog.LevelDebug, "received", slog.Any("event", wrObj))

		c.handleMessage(ctx, &wrObj)
	}
}

// handleMessage received from the Omlox Hub server.
func (c *Client) handleMessage(ctx context.Context, msg *wrapperObject) {
	switch msg.Event {
	case EventError:
		c.handleError(ctx, msg)
	case EventSubscribed:
		// pop pending subscription and assign subscription ID
		pendingc := <-c.pending
		chsend(ctx, pendingc, struct {
			sid int
			err error
		}{
			sid: msg.SubscriptionID,
		})
		return
	case EventUnsubscribed:
		// TODO @dvcorreia: close subscription
	default:
		c.routeMessage(ctx, &msg.WrapperObject)
	}
}

// handleError handles any websocket error sent by the server.
func (c *Client) handleError(ctx context.Context, msg *wrapperObject) {
	switch msg.WebsocketError.Code {
	case ErrCodeSubscription, ErrCodeNotAuthorized, ErrCodeUnknownTopic, ErrCodeInvalid:
		// pop pending subscription and kill it
		pendingc := <-c.pending
		chsend(ctx, pendingc, struct {
			sid int
			err error
		}{
			err: msg.WebsocketError,
		})
		return
	case ErrCodeUnknown: // TODO @dvcorreia: handle error
	case ErrCodeUnsubscription: // TODO @dvcorreia: handle error
	}
}

// routeMessage sends the message to the its respective subscription.
func (c *Client) routeMessage(ctx context.Context, msg *WrapperObject) {
	// retrive subcription if exists
	c.mu.RLock()
	sub := c.subs[msg.SubscriptionID]
	c.mu.RUnlock()

	if sub == nil {
		// TODO @dvcorreia: handle unknown subscription IDs
		return
	}

	select {
	case <-ctx.Done():
		return
	case sub.mch <- msg: // TODO @dvcorreia: this will block other messages
	case <-time.After(chanSendTimeout):
		slog.LogAttrs(
			context.Background(),
			slog.LevelWarn,
			"timeout sending to subscription channel",
			slog.Any("event", msg),
		)
	}
}

// clearSubs closes resources of subscriptions.
func (c *Client) clearSubs() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for sid, sub := range c.subs {
		sub.close()
		delete(c.subs, sid)
	}

	// close any pending subscription
	select {
	case pending := <-c.pending:
		pending <- struct {
			sid int
			err error
		}{
			err: net.ErrClosed,
		}
	default:
	}

	close(c.pending)
}

// Close releases any resources held by the client,
// such as connections, memory and goroutines.
func (c *Client) Close() error {
	if !c.isClosed() {
		err := c.conn.Close(websocket.StatusNormalClosure, "")
		if err != nil {
			return err
		}
	}

	// close the client context
	c.cancel()

	c.lifecycleWg.Wait()

	return nil
}

// isClosed reports if the client closed.
func (c *Client) isClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

func upgradeToWebsocketScheme(u *url.URL) error {
	switch u.Scheme {
	case httpScheme:
		u.Scheme = wsScheme
	case httpSchemeTLS:
		u.Scheme = wsSchemeTLS
	default:
		return fmt.Errorf("invalid websocket scheme '%s'", u.Scheme)
	}

	return nil
}

// LogValue implements slog.LogValuer.
func (w wrapperObject) LogValue() slog.Value {
	if w.Event == EventError {
		return slog.GroupValue(
			slog.Any("error", w.WebsocketError),
		)
	}
	return w.WrapperObject.LogValue()
}

// chsend sends a value to channel with context cancelation.
func chsend[T any](ctx context.Context, ch chan T, v T) {
	select {
	case <-ctx.Done():
		return
	case ch <- v:
	}
}
