// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// Connect dials the Omlox™ Hub websockets interface with automatic retry support.
func (c *Client) Connect(ctx context.Context) error {
	var err error
	var shouldRetry bool
	var attempt int

	if c.configuration.WSAutoReconnect {
		c.mu.Lock()
		c.reconnectCtx, c.reconnectCancel = context.WithCancel(context.Background())
		c.mu.Unlock()
	}

	for attempt = 0; ; attempt++ {
		err = c.connect(ctx)
		if err == nil {
			return nil
		}

		shouldRetry, err = c.configuration.WSCheckRetry(ctx, attempt, err)
		if !shouldRetry {
			break
		}

		remain := c.configuration.WSMaxRetries - attempt
		if c.configuration.WSMaxRetries >= 0 && remain <= 0 {
			break
		}

		wait := c.configuration.WSBackoff(
			c.configuration.WSMinRetryWait,
			c.configuration.WSMaxRetryWait,
			attempt,
		)

		slog.LogAttrs(
			ctx,
			slog.LevelDebug,
			"retrying connection",
			slog.Int("attempt", attempt+1),
			slog.Duration("wait", wait),
			slog.Int("remaining", remain),
			slog.Any("error", err),
		)

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	// all retries exhausted
	if err == nil {
		return fmt.Errorf("connection failed after %d attempt(s)", attempt)
	}
	return fmt.Errorf("connection failed after %d attempt(s): %w", attempt, err)
}

// connect performs a single connection attempt to the Omlox™ Hub websockets interface.
func (c *Client) connect(ctx context.Context) error {
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
	c.lifecycleWg.Add(2)
	c.cancel = cancel
	c.mu.Unlock()

	go func() {
		defer c.lifecycleWg.Done()
		if err := c.readLoop(ctx, conn); err != nil {
			c.handleConnectionLoss(err)
		}
	}()

	go func() {
		defer c.lifecycleWg.Done()
		if err := c.pingLoop(ctx, conn); err != nil {
			c.handleConnectionLoss(err)
		}
	}()

	return nil
}

// handleConnectionLoss handles unexpected connection failures and triggers reconnection.
// This is called by readLoop or pingLoop when they detect a connection error.
//
// Reconnection is performed asynchronously in a new goroutine so that the
// calling goroutine (readLoop/pingLoop) can return and release its WaitGroup
// slot. The reconnection goroutine waits for both old goroutines to exit
// before dialing a new connection.
func (c *Client) handleConnectionLoss(err error) {
	c.mu.Lock()

	// Prevent multiple concurrent reconnection attempts. Both readLoop and
	// pingLoop may detect the failure and call this method.
	if c.reconnecting {
		c.mu.Unlock()
		return
	}

	autoReconnect := c.configuration.WSAutoReconnect
	reconnectCtx := c.reconnectCtx

	if !autoReconnect {
		c.mu.Unlock()
		slog.LogAttrs(
			context.Background(),
			slog.LevelWarn,
			"connection lost, auto-reconnect disabled",
			slog.Any("error", err),
		)
		return
	}

	// client is shutting down, do not attempt reconnect
	if reconnectCtx.Err() != nil {
		c.mu.Unlock()
		return
	}

	c.reconnecting = true
	cancel := c.cancel
	c.mu.Unlock()

	slog.LogAttrs(
		context.Background(),
		slog.LevelWarn,
		"connection lost, attempting to reconnect",
		slog.Any("error", err),
	)

	// Cancel the old connection context so the other goroutine
	// (readLoop or pingLoop) stops and releases its WaitGroup slot.
	if cancel != nil {
		cancel()
	}

	// Reconnect asynchronously so the calling goroutine can return and
	// call lifecycleWg.Done(). The reconnection goroutine waits for both
	// old goroutines to exit before dialing a new connection.
	go func() {
		// Wait for old readLoop/pingLoop goroutines to finish.
		c.lifecycleWg.Wait()

		// Drain any stale pending subscription that was in-flight
		// when the connection was lost.
		c.drainPending()

		// Verify the client wasn't closed while we were waiting.
		if reconnectCtx.Err() != nil {
			c.mu.Lock()
			c.reconnecting = false
			c.mu.Unlock()
			return
		}

		if err := c.reconnect(reconnectCtx); err != nil {
			slog.LogAttrs(
				context.Background(),
				slog.LevelError,
				"reconnection failed",
				slog.Any("error", err),
			)
		} else {
			slog.LogAttrs(
				context.Background(),
				slog.LevelInfo,
				"successfully reconnected",
			)
		}

		c.mu.Lock()
		c.reconnecting = false
		c.mu.Unlock()
	}()
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

	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()

	if closed {
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

	sub := &Subcription{
		topic:  topic,
		params: parameters,
		mch:    make(chan *WrapperObject, receiveChanSize),
	}

	// perform the subscription handshake. the subscription is registered by
	// the confirmation handler, which also assigns its id.
	if err := c.subscribe(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// Sends a subscription message and handles the confirmation from the server.
//
// The subscription will be attributed an ID that can used for futher context.
// There can only be one pending subscription at each time.
// Subsequent subscriptions will wait while the pending one is waiting for an ID from the server.
// Since each subscription on a topic can have a distinct parameters, we must synchronisly wait to match each one to its ID.
func (c *Client) subscribe(ctx context.Context, sub *Subcription) error {
	pending := pendingSubscription{
		sub: sub,
		// channel to await subscription confirmation
		await: make(chan subscribeResult),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	// lock for pending subscription confirmation.
	// the pending will be freed by the subribed message handler.
	case c.pending <- pending:
	}

	wrObj := &WrapperObject{
		Event:   EventSubscribe,
		Topic:   sub.topic,
		Params:  sub.params,
		Payload: nil,
	}

	if err := c.publish(ctx, wrObj); err != nil {
		select {
		case <-c.pending:
		default:
		}
		return err
	}

	// wait for subcription ID
	var r subscribeResult
	select {
	case <-ctx.Done():
		return ctx.Err()
	case r = <-pending.await:
	}

	return r.err
}

// ping pong loop that manages the websocket connection health.
func (c *Client) pingLoop(ctx context.Context, conn *websocket.Conn) error {
	t := time.NewTicker(pingPeriod)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}

		// Use a dedicated variable to avoid shadowing the parent ctx.
		// Shadowing would cause each iteration to derive from the previous
		// timeout context, compounding the deadline until pings fail.
		pingCtx, cancel := context.WithTimeout(ctx, pongWait)
		begin := time.Now()
		err := conn.Ping(pingCtx)
		cancel() // release immediately instead of deferring in a loop

		if err != nil {
			// Intentional shutdown — do not reconnect.
			if errors.Is(err, context.Canceled) {
				return nil
			}

			// Ping failed (deadline exceeded, connection error, etc.) —
			// return error to trigger reconnection.
			return err
		}

		slog.Debug("heartbeat", slog.Duration("latency", time.Since(begin)))
	}
}

// readLoop that will handle incomming data.
func (c *Client) readLoop(ctx context.Context, conn *websocket.Conn) error {
	// set the client to closed state when this goroutine exits
	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		// marked closed if it is the active connection
		if c.conn == conn {
			c.closed = true
		}
	}()

	for {
		msgType, r, err := conn.Reader(ctx)

		if err != nil {
			// Intentional shutdown via context cancellation — do not reconnect.
			if errors.Is(err, context.Canceled) {
				return nil
			}

			// Check for clean WebSocket close frames before other error types.
			// StatusNormalClosure is an intentional close — do not reconnect.
			// StatusGoingAway means the server is shutting down (e.g. restart) — reconnect.
			switch websocket.CloseStatus(err) {
			case websocket.StatusNormalClosure:
				return nil
			case websocket.StatusGoingAway:
				return err
			}

			// All other errors (io.EOF from dropped TCP connections, net.Error
			// from timeouts/resets, unexpected close codes) indicate the
			// connection was lost and should trigger reconnection.
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

// reconnect attempts to re-establish the WebSocket connection with retry logic.
func (c *Client) reconnect(ctx context.Context) error {
	var err error
	var shouldRetry bool
	var attempt int

	for attempt = 0; ; attempt++ {
		err = c.connect(ctx)
		if err == nil {
			// connection successful, restore subscriptions
			return c.restoreSubscriptions(ctx)
		}

		shouldRetry, err = c.configuration.WSCheckRetry(ctx, attempt, err)
		if !shouldRetry {
			break
		}

		remain := c.configuration.WSMaxRetries - attempt
		if c.configuration.WSMaxRetries >= 0 && remain <= 0 {
			break
		}

		wait := c.configuration.WSBackoff(
			c.configuration.WSMinRetryWait,
			c.configuration.WSMaxRetryWait,
			attempt,
		)

		slog.LogAttrs(
			ctx,
			slog.LevelDebug,
			"retrying reconnection",
			slog.Int("attempt", attempt+1),
			slog.Duration("wait", wait),
			slog.Int("remaining", remain),
			slog.Any("error", err),
		)

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	// all retries exhausted
	if err == nil {
		return fmt.Errorf("reconnection failed after %d attempt(s)", attempt)
	}
	return fmt.Errorf("reconnection failed after %d attempt(s): %w", attempt, err)
}

// restoreSubscriptions re-establishes all active subscriptions after reconnection.
func (c *Client) restoreSubscriptions(ctx context.Context) error {
	c.mu.RLock()
	// snapshot the subscriptions to avoid holding the lock during network calls
	subsToRestore := make([]*Subcription, 0, len(c.subs))
	for _, sub := range c.subs {
		subsToRestore = append(subsToRestore, sub)
	}
	c.mu.RUnlock()

	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"restoring subscriptions",
		slog.Int("count", len(subsToRestore)),
	)

	// clear old subscriptions map since ids will change.
	// subscriptions will be repopulated in resubscription
	c.mu.Lock()
	c.subs = make(map[int]*Subcription)
	c.mu.Unlock()

	for _, sub := range subsToRestore {
		oldSid := sub.sid

		// the same subscription object is re-registered under its new sid by
		// the confirmation handler, so the user keeps reading the same channel
		if err := c.subscribe(ctx, sub); err != nil {
			slog.LogAttrs(
				ctx,
				slog.LevelError,
				"failed to restore subscription",
				slog.Int("old sid", oldSid),
				slog.String("topic", string(sub.topic)),
				slog.Any("error", err),
			)
			continue
		}

		slog.LogAttrs(
			ctx,
			slog.LevelDebug,
			"subscription restored",
			slog.Int("old sid", oldSid),
			slog.Int("new sid", sub.sid),
			slog.String("topic", string(sub.topic)),
		)
	}

	return nil
}

// handleMessage received from the Omlox Hub server.
func (c *Client) handleMessage(ctx context.Context, msg *wrapperObject) {
	switch msg.Event {
	case EventError:
		c.handleError(ctx, msg)
	case EventSubscribed:
		// pop pending subscription, assign its subscription ID and register it.
		//
		// registering here, on the goroutine reading the connection, is what
		// guarantees the subscription is routable before the next frame is
		// handled — hubs start pushing as soon as they confirm.
		select {
		case pending := <-c.pending:
			c.mu.Lock()
			pending.sub.sid = msg.SubscriptionID
			c.subs[msg.SubscriptionID] = pending.sub
			c.mu.Unlock()

			chsend(ctx, pending.await, subscribeResult{sid: msg.SubscriptionID})
		default:
			slog.Warn("received subscription confirmation but no pending subscription found")
		}
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
		pending := <-c.pending
		chsend(ctx, pending.await, subscribeResult{err: msg.WebsocketError})
		return
	case ErrCodeUnknown: // TODO @dvcorreia: handle error
	case ErrCodeUnsubscription: // TODO @dvcorreia: handle error
	}
}

// routeMessage sends the message to the its respective subscription.
func (c *Client) routeMessage(ctx context.Context, msg *WrapperObject) {
	subs := c.resolveSubscriptions(msg)

	if len(subs) == 0 {
		slog.LogAttrs(
			context.Background(),
			slog.LevelWarn,
			"dropping message with no matching subscription",
			slog.Any("event", msg),
		)
		return
	}

	for _, sub := range subs {
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
}

// resolveSubscriptions finds the subscriptions a message belongs to.
//
// The omlox™ specification has the hub echo the subscription_id that generated
// the data on every message, and that is the only unambiguous way to route a
// message when the same topic is subscribed more than once with different
// parameters. Some hubs (Deephub) only return the subscription_id on the
// subscribed confirmation and omit it from subsequent messages, which decodes
// to the zero value here. For those, fall back to matching on the topic.
func (c *Client) resolveSubscriptions(msg *WrapperObject) []*Subcription {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if sub := c.subs[msg.SubscriptionID]; sub != nil {
		return []*Subcription{sub}
	}

	if msg.Topic == "" {
		return nil
	}

	var subs []*Subcription
	for _, sub := range c.subs {
		if sub.topic == msg.Topic {
			subs = append(subs, sub)
		}
	}

	return subs
}

// drainPending clears any stale pending subscription from the channel,
// notifying the waiter about the closed connection.
func (c *Client) drainPending() {
	select {
	case stale := <-c.pending:
		select {
		case stale.await <- subscribeResult{err: net.ErrClosed}:
		default:
		}
	default:
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

	// Drain any pending subscription and reinitialize the channel.
	// We reinitialize rather than closing to prevent send-on-closed-channel
	// panics if a reconnection goroutine races with Close().
	select {
	case stale := <-c.pending:
		select {
		case stale.await <- subscribeResult{err: net.ErrClosed}:
		default:
		}
	default:
	}

	c.pending = make(chan pendingSubscription, 1)
}

// Close releases any resources held by the client,
// such as connections, memory and goroutines.
func (c *Client) Close() error {
	// stop automatic reconnection
	c.mu.Lock()
	if c.reconnectCancel != nil {
		c.reconnectCancel()
	}
	c.mu.Unlock()

	if !c.isClosed() {
		err := c.conn.Close(websocket.StatusNormalClosure, "")
		if err != nil {
			return err
		}
	}

	// cancel context to stop goroutines
	if c.cancel != nil {
		c.cancel()
	}

	// wait for all goroutines to finish
	c.lifecycleWg.Wait()

	// clear subscriptions after all goroutines have stopped
	c.clearSubs()

	return nil
}

// isClosed reports if the client closed.
func (c *Client) isClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// IsConnected reports if the client is currently connected to the WebSocket.
// Returns true if connected, false if closed or reconnecting.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.closed && !c.reconnecting
}

// IsReconnecting reports if the client is currently attempting to reconnect.
func (c *Client) IsReconnecting() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reconnecting
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
