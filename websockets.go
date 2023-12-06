// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"errors"

	"golang.org/x/exp/slog"
)

// WrapperObject is the wrapper object of websockets data exchanged between client and server.
//
//easyjson:json
type WrapperObject struct {
	// Event is always required for all data exchanged between client and server.
	Event Event `json:"event"`
	Topic Topic `json:"topic,omitempty"`

	// The concrete topic subscription which generated the data.
	SubscriptionID int `json:"subscription_id,omitempty"`

	// An array containing valid omlox™ data objects (or empty).
	Payload []json.RawMessage `json:"payload,omitempty"`

	// Optional object containing key-value pairs of parameters.
	// Parameters usually match their REST API counterparts.
	Params Parameters `json:"params,omitempty"`
}

var _ slog.LogValuer = (*WrapperObject)(nil)

// LogValue implements slog.LogValuer.
func (w WrapperObject) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("type", string(w.Event)),
		slog.String("topic", string(w.Topic)),
		slog.Int("sid", w.SubscriptionID),
		slog.Any("params", w.Params),
	)
}

// Parameter is an optional key-value pair used on subscriptions.
// If a parameter is unsupported for a certain reason, it must return an error.
type Parameter func(Topic, Parameters) error

// Parameters represents the key-value pairs used in subscriptions.
type Parameters map[string]string

var _ slog.LogValuer = (*Parameters)(nil)

func (p Parameters) LogValue() slog.Value {
	if p == nil {
		return slog.Value{}
	}

	logv := make([]slog.Attr, 0, len(p))

	for name, val := range p {
		logv = append(logv, slog.String(name, val))
	}

	return slog.GroupValue(logv...)
}

// WebsocketError sent to the client on websocket server error.
//
//easyjson:json
type WebsocketError struct {
	Code        ErrCode `json:"code,omitempty"`
	Description string  `json:"description,omitempty"`
}

var (
	_ error          = (*WebsocketError)(nil)
	_ slog.LogValuer = (*WebsocketError)(nil)
)

func (err WebsocketError) Error() string {
	return err.Code.String() + ": " + err.Description
}

func (err WebsocketError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("code", int(err.Code)),
		slog.String("description", err.Description),
	)
}

// event abstracts the possible events types in websocket messages.
type Event string

const (
	EventMsg          Event = "message"
	EventSubscribe    Event = "subscribe"
	EventSubscribed   Event = "subscribed"
	EventUnsubscribe  Event = "unsubscribe"
	EventUnsubscribed Event = "unsubscribed"
	EventError        Event = "error"
)

// UnmarshalJSON decodes type from JSON.
func (e *Event) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// options should match the event const option types.
	switch wse := Event(s); wse {
	case EventMsg,
		EventSubscribe,
		EventSubscribed,
		EventUnsubscribe,
		EventUnsubscribed,
		EventError:
		*e = wse
	default:
		return errors.New("unsuported websocket event type")
	}

	return nil
}

// Omlox Hub supports a a few topics to which clients can subscribe and publish.
type Topic string

const (
	// For real-time location updates, as well as sending location updates to the hub.
	// When receiving data for this topic the payload of the wrapper object contains omlox™ Location objects.
	TopicLocationUpdates Topic = "location_updates"

	// To retrieve location information as GeoJson feature collection.
	TopicLocationUpdatesGeoJSON Topic = "location_updates:geojson"

	// Checks trackable movements for collisions and sends collision events when trackables:
	// start to collide, continue to collide and end a collision.
	// When receiving data for this topic the payload of the wrapper object contains omlox™ CollisionEvent objects.
	TopicCollisionEvents Topic = "collision_events"

	// To inform subscribers about geofence entry and exit events.
	// When receiving data for this topic the payload of the wrapper object contains omlox™ FenceEvent objects.
	TopicFenceEvents Topic = "fence_events"

	// Similar to fence events, but instead of an omlox™ FenceEvent object GeoJson feature collections are returned as payload.
	TopicFenceEventsGeoJSON Topic = "fence_events:geojson"

	// To receive movements of omlox™ Trackables.
	// When receiving data for this topic, the payload of the wrapper object contains omlox™ TrackableMotion objects.
	TopicTrackableMotions Topic = "trackable_motions"
)

// ErrCode is an error code used by applications to discern the type of the websocket error.
type ErrCode int

const (
	// Event type is unknown.
	ErrCodeUnknown ErrCode = 10000

	// Unknown topic name.
	ErrCodeUnknownTopic ErrCode = 10001

	// Subscription failed.
	ErrCodeSubscription ErrCode = 10002

	// Unsubscribe failed.
	ErrCodeUnsubscription ErrCode = 10003

	// Not authorized.
	ErrCodeNotAuthorized ErrCode = 10004

	// Invalid payload data.
	ErrCodeInvalid ErrCode = 10005
)

// Map between error codes and their text representation.
// It is public to facilitate adding custom error codes for client extensions.
var ErrCodeMap = map[ErrCode]string{
	ErrCodeUnknown:        "event type is unknown",
	ErrCodeUnknownTopic:   "unknown topic name",
	ErrCodeSubscription:   "subscription failed",
	ErrCodeUnsubscription: "unsubscribe failed",
	ErrCodeNotAuthorized:  "not authorized",
	ErrCodeInvalid:        "invalid payload data",
}

// String return a text representation.
func (e ErrCode) String() string {
	str, ok := ErrCodeMap[e]
	if !ok {
		return "event type is unknown"
	}
	return str
}
