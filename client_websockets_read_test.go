// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

// pushingHubServer is an Omlox Hub websocket interface that acknowledges a
// subscription and then writes each of frames, as raw text messages, on it.
func pushingHubServer(t *testing.T, frames ...[]byte) *httptest.Server {
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

			if msg.Event != EventSubscribe {
				continue
			}

			ack := WrapperObject{Event: EventSubscribed, Topic: msg.Topic, SubscriptionID: 1}
			if err := writeWrapper(ctx, conn, ack); err != nil {
				return
			}

			for _, frame := range frames {
				if err := conn.Write(ctx, websocket.MessageText, frame); err != nil {
					return
				}
			}
		}
	}))

	t.Cleanup(srv.Close)

	return srv
}

// locationMessage returns a location_updates message whose payload holds n
// copies of the routing test location.
func locationMessage(t *testing.T, n int) []byte {
	t.Helper()

	payload := make([]json.RawMessage, n)
	for i := range payload {
		payload[i] = json.RawMessage(routingTestPayload)
	}

	data, err := json.Marshal(WrapperObject{
		Event:          EventMsg,
		Topic:          TopicLocationUpdates,
		SubscriptionID: 1,
		Payload:        payload,
	})
	if err != nil {
		t.Fatalf("marshalling message: %v", err)
	}

	return data
}

// receive waits for the next message on sub, failing the test on a closed
// channel or a timeout.
func receive(t *testing.T, sub *Subcription) *WrapperObject {
	t.Helper()

	select {
	case msg := <-sub.ReceiveRaw():
		if msg == nil {
			t.Fatal("subscription channel closed")
		}
		return msg
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for a message")
	}

	return nil
}

// TestReadLargeMessage guards the read limit. The WebSocket library's default
// of 32 KiB is below the size of the messages a busy hub sends, and a message
// over the limit closed the connection, so the client dropped every such
// message and reconnected.
func TestReadLargeMessage(t *testing.T) {
	t.Parallel()

	large := locationMessage(t, 400)
	if len(large) <= 32768 {
		t.Fatalf("test message is %d bytes, want more than 32 KiB", len(large))
	}

	srv := pushingHubServer(t, large, locationMessage(t, 1))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := Connect(ctx, srv.URL+"/v2")
	if err != nil {
		t.Fatalf("connecting to hub: %v", err)
	}
	defer client.Close()

	sub, err := client.Subscribe(ctx, TopicLocationUpdates)
	if err != nil {
		t.Fatalf("subscribing: %v", err)
	}

	if got := len(receive(t, sub).Payload); got != 400 {
		t.Errorf("first message: payload length = %d, want 400", got)
	}
	if got := len(receive(t, sub).Payload); got != 1 {
		t.Errorf("second message: payload length = %d, want 1", got)
	}
	if !client.IsConnected() {
		t.Error("client is not connected after reading a large message")
	}
}

// TestReadMessageOverLimit checks that a message over the configured read
// limit is handled as a lost connection, rather than skipped as if it were a
// malformed message while the connection is already closed.
func TestReadMessageOverLimit(t *testing.T) {
	t.Parallel()

	srv := pushingHubServer(t, locationMessage(t, 10))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := Connect(ctx, srv.URL+"/v2", WithWSReadLimit(512))
	if err != nil {
		t.Fatalf("connecting to hub: %v", err)
	}
	defer client.Close()

	if _, err := client.Subscribe(ctx, TopicLocationUpdates); err != nil {
		t.Fatalf("subscribing: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for client.IsConnected() {
		if time.Now().After(deadline) {
			t.Fatal("client still connected after a message over the read limit")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestReadInvalidMessage checks that a message that is not valid JSON is
// dropped and the messages after it are still delivered on the same
// connection.
func TestReadInvalidMessage(t *testing.T) {
	t.Parallel()

	srv := pushingHubServer(t, []byte(`{"event":"message","payload":[`+strings.Repeat(" ", 64)), locationMessage(t, 1))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := Connect(ctx, srv.URL+"/v2")
	if err != nil {
		t.Fatalf("connecting to hub: %v", err)
	}
	defer client.Close()

	sub, err := client.Subscribe(ctx, TopicLocationUpdates)
	if err != nil {
		t.Fatalf("subscribing: %v", err)
	}

	if got := len(receive(t, sub).Payload); got != 1 {
		t.Errorf("payload length = %d, want 1", got)
	}
	if !client.IsConnected() {
		t.Error("client is not connected after an invalid message")
	}
}

func TestWithWSReadLimit(t *testing.T) {
	tests := []struct {
		limit   int64
		wantErr bool
	}{
		{limit: 1 << 20},
		{limit: 1},
		{limit: -1},
		{limit: 0, wantErr: true},
		{limit: -2, wantErr: true},
	}

	for _, tt := range tests {
		c := DefaultConfiguration()
		err := WithWSReadLimit(tt.limit)(&c)
		if (err != nil) != tt.wantErr {
			t.Errorf("WithWSReadLimit(%d) error = %v, want error %v", tt.limit, err, tt.wantErr)
			continue
		}
		if err == nil && c.WSReadLimit != tt.limit {
			t.Errorf("WithWSReadLimit(%d) set %d", tt.limit, c.WSReadLimit)
		}
	}

	if got := DefaultConfiguration().WSReadLimit; got != DefaultWSReadLimit {
		t.Errorf("default read limit = %d, want %d", got, DefaultWSReadLimit)
	}
}
