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

// hubDialect describes how a hub reports the subscription id on the messages
// it pushes after a subscription is confirmed.
type hubDialect struct {
	name string
	// subscribedID is the id returned on the `subscribed` confirmation.
	subscribedID int
	// echoID reports whether `message` frames carry the subscription_id.
	// The omlox™ specification requires it; Deephub omits it.
	echoID bool
}

var hubDialects = []hubDialect{
	// Spec compliant: the id assigned on subscribe is echoed on every message.
	{name: "echoes-subscription-id", subscribedID: 1, echoID: true},
	// Deephub: assigns an id, then omits it from every subsequent message.
	{name: "omits-subscription-id", subscribedID: 16727, echoID: false},
}

const routingTestPayload = `{"position":{"type":"Point","coordinates":[5,4]},"source":"fdb6df62-bce8-6c23-e342-80bd5c938774","provider_type":"uwb","provider_id":"77:4f:34:69:27:40","timestamp_generated":"2019-09-02T22:02:24.355Z","timestamp_sent":"2019-09-02T22:02:24.355Z"}`

// hubServer is a minimal Omlox Hub websocket interface that acknowledges
// subscriptions and then pushes a single message on the subscribed topic.
func hubServer(t *testing.T, dialect hubDialect) *httptest.Server {
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

			ack := WrapperObject{
				Event:          EventSubscribed,
				Topic:          msg.Topic,
				SubscriptionID: dialect.subscribedID,
			}
			if err := writeWrapper(ctx, conn, ack); err != nil {
				return
			}

			push := WrapperObject{
				Event:   EventMsg,
				Topic:   msg.Topic,
				Payload: []json.RawMessage{json.RawMessage(routingTestPayload)},
			}
			if dialect.echoID {
				push.SubscriptionID = dialect.subscribedID
			}
			if err := writeWrapper(ctx, conn, push); err != nil {
				return
			}
		}
	}))

	t.Cleanup(srv.Close)

	return srv
}

func writeWrapper(ctx context.Context, conn *websocket.Conn, msg WrapperObject) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return conn.Write(ctx, websocket.MessageText, data)
}

// TestSubscriptionRouting guards the routing of pushed messages onto the
// subscription that asked for them, against both hub dialects. Routing on a
// hardcoded subscription id of zero used to silently drop every message from a
// hub that echoes the real id.
func TestSubscriptionRouting(t *testing.T) {
	for _, dialect := range hubDialects {
		dialect := dialect

		t.Run(dialect.name, func(t *testing.T) {
			t.Parallel()

			srv := hubServer(t, dialect)

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

			if got, want := sub.sid, dialect.subscribedID; got != want {
				t.Errorf("subscription id = %d, want %d", got, want)
			}

			select {
			case msg := <-sub.ReceiveRaw():
				if msg == nil {
					t.Fatal("subscription channel closed")
				}
				if msg.Topic != TopicLocationUpdates {
					t.Errorf("topic = %q, want %q", msg.Topic, TopicLocationUpdates)
				}
				if len(msg.Payload) != 1 {
					t.Fatalf("payload length = %d, want 1", len(msg.Payload))
				}
			case <-time.After(5 * time.Second):
				t.Fatal("timed out waiting for a location update")
			}
		})
	}
}

// TestResolveSubscriptions covers the lookup in isolation, including the
// fallback used for hubs that omit the subscription id from their messages.
func TestResolveSubscriptions(t *testing.T) {
	locations := &Subcription{sid: 1, topic: TopicLocationUpdates}
	fences := &Subcription{sid: 2, topic: TopicFenceEvents}

	c := &Client{subs: map[int]*Subcription{
		locations.sid: locations,
		fences.sid:    fences,
	}}

	tests := []struct {
		name string
		msg  WrapperObject
		want []*Subcription
	}{
		{
			name: "by subscription id",
			msg:  WrapperObject{Topic: TopicLocationUpdates, SubscriptionID: 1},
			want: []*Subcription{locations},
		},
		{
			name: "id takes precedence over topic",
			msg:  WrapperObject{Topic: TopicLocationUpdates, SubscriptionID: 2},
			want: []*Subcription{fences},
		},
		{
			name: "falls back to topic when the id is absent",
			msg:  WrapperObject{Topic: TopicFenceEvents},
			want: []*Subcription{fences},
		},
		{
			name: "unknown topic without an id resolves to nothing",
			msg:  WrapperObject{Topic: TopicCollisionEvents},
			want: nil,
		},
		{
			name: "no topic and no id resolves to nothing",
			msg:  WrapperObject{},
			want: nil,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			got := c.resolveSubscriptions(&tc.msg)

			if len(got) != len(tc.want) {
				t.Fatalf("resolved %d subscriptions, want %d", len(got), len(tc.want))
			}

			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("subscription %d = %v, want %v", i, got[i], tc.want[i])
				}
			}
		})
	}
}
