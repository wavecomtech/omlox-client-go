// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

// rejectingHubServer is an Omlox Hub websocket interface that rejects every
// published message with an `invalid payload data` error, the way the hub does
// for a location it will not accept. It still acknowledges subscriptions, so a
// subscribe can be exercised after a rejection.
func rejectingHubServer(t *testing.T) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusInternalError, "")

		ctx := r.Context()

		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}

			var msg WrapperObject
			if err := json.Unmarshal(data, &msg); err != nil {
				return
			}

			switch msg.Event {
			case EventMsg:
				// the hub reports a rejected publication as a bare error
				// frame: no topic, no subscription id to correlate it with.
				payload, err := json.Marshal(map[string]any{
					"event":       EventError,
					"code":        ErrCodeInvalid,
					"description": "invalid location payload",
				})
				if err != nil {
					return
				}
				if err := conn.Write(ctx, websocket.MessageText, payload); err != nil {
					return
				}
			case EventSubscribe:
				ack := WrapperObject{
					Event:          EventSubscribed,
					Topic:          msg.Topic,
					SubscriptionID: 1,
				}
				if err := writeWrapper(ctx, conn, ack); err != nil {
					return
				}
			}
		}
	}))

	t.Cleanup(srv.Close)

	return srv
}

// TestPublishRejectionDoesNotWedgeClient guards the read loop against an error
// frame that belongs to no pending subscription. Popping the pending
// subscription unconditionally used to block the read loop forever, which
// wedged Close() with it — and a client that only publishes, as the RTLS
// adapters do, never has a pending subscription for such an error to claim.
func TestPublishRejectionDoesNotWedgeClient(t *testing.T) {
	srv := rejectingHubServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := Connect(ctx, srv.URL+"/v2")
	if err != nil {
		t.Fatalf("connecting to hub: %v", err)
	}

	if err := client.Publish(ctx, TopicLocationUpdates, json.RawMessage(`{}`)); err != nil {
		t.Fatalf("publishing: %v", err)
	}

	// give the rejection time to be read and handled
	time.Sleep(500 * time.Millisecond)

	// the client must still be usable: the subscribe handshake needs the very
	// pending slot the rejection would have consumed.
	if _, err := client.Subscribe(ctx, TopicLocationUpdates); err != nil {
		t.Errorf("subscribing after a rejected publication: %v", err)
	}

	closed := make(chan error, 1)
	go func() { closed <- client.Close() }()

	select {
	case <-closed:
	case <-time.After(10 * time.Second):
		t.Fatal("Close() blocked after a rejected publication: the read loop is wedged")
	}
}
