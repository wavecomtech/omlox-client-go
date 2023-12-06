// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Trackable defines model for Trackable.
//
//easyjson:json
type Trackable struct {
	// Must be a UUID. When creating a trackable, a unique id will be generated if it is not provided.
	ID uuid.UUID `json:"id"`

	// Either 'omlox' or 'virtual'. An omlox™ compatible trackable has knowledge of it's location providers
	// (e.g. embedded UWB, BLE, RFID hardware), and self-assigns it's location providers.
	// A virtual trackable can be used to assign location providers to a logical asset.
	Type TrackableType `json:"type"`

	// A describing name
	Name string `json:"name,omitempty"`

	// GeoJson Polygon geometry. Important: A Polygon object MUST be interpreted according to a coordinate reference system (crs).
	// The ordering of components is x,y,z or longitude,latitude,elevation respectively as according to the GeoJson specification.
	Geometry *Polygon `json:"geometry,omitempty"`

	// The extrusion to be applied to the geometry in meters.
	// Must be a positive number.
	Extrusion float64 `json:"extrusion,omitempty"`

	// The location provider ids (e.g. mac addresses) assigned to this trackable.
	// Note: An application may create virtual location providers and assign these to a trackable where desired.
	// This allows applications to identify trackables for location providers which themselve do not have a
	// unique identifier (e.g. certain GPS devices).
	LocationProviders []string `json:"location_providers,omitempty"`

	// The timeout in milliseconds after which a location should expire and optional
	// trigger a fence exit event (if no more location updates are sent).
	// Must be a positive number or -1 in case of an infinite timeout.
	// If not set, or set to null, it will default to the fence setting.
	FenceTimeout Duration `json:"fence_timeout,omitempty"`

	// The minimum distance in meters for a trackable to release from an ongoing collision.
	// For example, for a trackable that was previously colliding with another trackable by being inside a trackable's radius, the collision
	// event will not be released from the collision until its distance to the trackable's geometry is at least the given exit_tolerance.
	// Must be a positive number. If not set, or set to null, it will default to the fence setting.
	ExitTolerance float64 `json:"exit_tolerance,omitempty"`

	// The timeout in milliseconds after which collision outside of a trackable but still within exit_tolerance distance to another
	// obstacle should release from a collision.
	// Must be a positive number or -1 in case of an infinite timeout.
	// If not set, or set to null, it will default to the fence setting.
	ToleranceTimeout Duration `json:"tolerance_timeout,omitempty"`

	// The delay in milliseconds in which an imminent exit event should wait for another location update.
	// This is relevant for fast rate position updates with quickly moving objects.
	// For example, an RTLS provider may batch location updates into groups, resulting in distances being temporarily outdated and
	// premature events between quickly moving objects.
	// The provided number must be positive or -1 in case of an infinite exit_delay.
	// If not set, or set to null, it will default to the fence setting.
	ExitDelay Duration `json:"exit_delay,omitempty"`

	// A radius provided in meters, defining the approximate circumference of the trackable.
	// If a radius value is set, all position updates from any of the Location Providers will generate a circular geometry
	// for the trackable, where the position is the center and the circle will be generated with the given radius.
	Radius float64 `json:"radius,omitempty"`

	// Any additional application or vendor specific properties.
	// An application implementing this object is not required to interpret any of the custom properties,
	// but it MUST preserve the properties if set.
	Properties json.RawMessage `json:"properties,omitempty"`

	// When a location update is processed, the locating rules of a trackable are applied to all its associated locations,
	// to determine its most significant location:
	// If a Boolean expression evaluates to true, the priority for the expression is applied to the location.
	// If multiple expressions evaluate to true, the highest priority is applied.
	// The location with the highest priority is considered the most significant location of that trackable.
	// If multiple locations share the highest priority, the most recent of these locations is the most significant.
	LocatingRules []LocatingRule `json:"locating_rules,omitempty"`
}

// Either 'omlox' or 'virtual'. An omlox™ compatible trackable has knowledge of it's location providers
// (e.g. embedded UWB, BLE, RFID hardware), and self-assigns it's location providers.
// A virtual trackable can be used to assign location providers to a logical asset.
type TrackableType int

// Defines values for TrackableType.
const (
	TrackableTypeOmlox TrackableType = iota
	TrackableTypeVirtual
)

// FromString assigs itself from type name.
func (t *TrackableType) FromString(name string) error {
	v, ok := map[string]TrackableType{
		TrackableTypeOmlox.String():   TrackableTypeOmlox,
		TrackableTypeVirtual.String(): TrackableTypeVirtual,
	}[name]

	if !ok {
		return fmt.Errorf("trackable of type %s not supported", name)
	}

	*t = v
	return nil
}

// String return a text representation.
func (t TrackableType) String() string {
	types := [...]string{
		"omlox",
		"virtual",
	}

	if len(types) < int(t) {
		return ""
	}

	return types[t]
}

// MarshalJSON encodes type in to JSON.
func (t TrackableType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON decodes type from JSON.
func (t *TrackableType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	return t.FromString(s)
}
