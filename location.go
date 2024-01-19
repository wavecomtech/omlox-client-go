// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Location defines model for Location.
//
//easyjson:json
type Location struct {
	// A GeoJson Point geometry. Describes a point (e.g. a position) in 2 or 3 dimensions. Important: A Point object MUST be interpreted
	// according to a coordinate reference system (crs), e.g. the crs field in the Location object. In practice this means a Position
	// object for example contains relative x,y,z data from a UWB system given in the coordinate system of a local zone and a
	// transformation needs to be applied to convert the position to a geographic coordinate. But the Position might also contain a
	// geographic coordinate (longitude, latitude) for example from a GPS system (where no coordinate transformation is needed), or
	// position data in projection of a UTM zone (where again transformation is required). The ordering of components is x,y,z or
	// longitude,latitude,elevation respectively as according to the GeoJson specification.
	Position Point `json:"position"`

	// Represents the unique identifier of the RTLS system (zone_id or foreign_id) which generated this location object, or the id of
	// a self-localizing device (e.g. a UWB tag / provider_id in GPS mode).
	Source string `json:"source"`

	// The location provider type which triggered this location update.
	ProviderType LocationProviderType `json:"provider_type"`

	// The location provider unique identifier, e.g. the mac address of a UWB location provider.
	ProviderID string `json:"provider_id"`

	// The ids of trackables the provider is assigned to.
	Trackables []uuid.UUID `json:"trackables,omitempty"`

	// The timestamp when the location was calculated.
	// The timestamp MUST be an ISO 8601 timestamp using UTC timezone and it SHOULD have millisecond precision to allow for precise
	// speed and course calculations. If no timestamp is provided, the hub will use its local current time.
	TimestampGenerated *time.Time `json:"timestamp_generated,omitempty"`

	// The timestamp when the location was sent over the network.
	// The optional timestamp MUST be an ISO 8601 timestamp using UTC timezone and it SHOULD have millisecond precision.
	// Note: No delivery guarantee is made in case the data is lost in transit.
	TimestampSent *time.Time `json:"timestamp_sent,omitempty"`

	// A projection identifier defining the projection of the provided location coordinate. The crs MUST be either a valid EPSG identifier
	// (https://epsg.io) or 'local' if the locations provided as a relative coordinate of the floor plan. For best interoperability and
	// worldwide coverage WGS84 (EPSG:4326) SHOULD be the preferred projection (as used also by GPS).
	// If the crs field is not present, 'local' MUST be assumed as the default.
	Crs string `json:"crs,omitempty"`

	// Whether a client is currently associated to a network. This property SHOULD be set if available for WiFi based positioning.
	Associated bool `json:"associated,omitempty"`

	// The horizontal accuracy of the location update in meters.
	Accuracy *float64 `json:"accuracy,omitempty"`

	// A logical and non-localized representation for a building floor. Floor 0 represents the floor designated as 'ground'.
	// Negative numbers designate floors below the ground floor and positive to indicate floors above the ground floor.
	// When implemented the floor value MUST match described logical numbering scheme, which can be different from any numbering used
	// within a building. Values can be expressed as an integer value, or as a float as required for mezzanine floor levels.
	Floor float64 `json:"floor,omitempty"`

	// An accurate orientation reading from a 'true' heading direction. The 'true' magnetic as opposed to the normal magnetic heading.
	// Applications SHOULD prefer the true heading if available. An invalid or currently unavailable heading MUST be indicated by
	// a negative value.
	TrueHeading *float64 `json:"true_heading,omitempty"`

	// The magnetic heading direction, which deviates from the true heading by a few degrees and differs slightly depending on
	// the location on the globe.
	MagneticHeading *float64 `json:"magnetic_heading,omitempty"`

	// The maximum deviation in degrees between the reported heading and the true heading.
	HeadingAccuracy *float64 `json:"heading_accuracy,omitempty"`

	// An elevation reference hint for the position's z component. If present it MUST be either 'floor' or 'wgs84'. If set to 'floor'
	// the z component MUST be assumed to be relative to the floor level. If set to 'wgs84' the z component MUST be treated as WGS84
	// ellipsoidal height. For the majority of applications an accurate geographic height may not be available. Therefore elevation_ref
	// MUST be assumed 'floor' by default if this property is not present.
	ElevationRef *ElevationRefType `json:"elevation_ref,omitempty"`

	// The current speed in meters per second. If the value is null or the property is not set, the current speed MAY be approximated
	// by an omlox™ hub based on the timestamp_generated value of a previous location update.
	Speed *float64 `json:"speed,omitempty"`

	// The current course ("compass direction"), which is the direction measured clockwise as an angle from true north on a compass
	// (north is 0°, east is 90°, south is 180°, and west is 270°). If the value is null or the property not set the course will be
	// approximated by an omlox™ hub based on the previous location.
	Course *float64 `json:"course,omitempty"`

	// Any additional application or vendor specific properties. An application implementing this  object is not required to interpret
	// any of the custom properties, but it MUST preserve the properties if set.
	Properties json.RawMessage `json:"properties,omitempty"`
}

// An elevation reference hint for the position's z component. Must be either 'floor' or 'wgs84'.
type ElevationRefType int

// Defines values for ElevationRefType.
const (
	ElevationRefTypeFloor ElevationRefType = iota
	ElevationRefTypeWgs84
)

// FromString assigs itself from type name.
func (e *ElevationRefType) FromString(name string) error {
	v, ok := map[string]ElevationRefType{
		ElevationRefTypeFloor.String(): ElevationRefTypeFloor,
		ElevationRefTypeWgs84.String(): ElevationRefTypeWgs84,
	}[name]

	if !ok {
		return fmt.Errorf("elevation reference of type %s not supported", name)
	}

	*e = v
	return nil
}

// String return a text representation.
func (e ElevationRefType) String() string {
	refs := [...]string{"floor", "wgs84"}
	if len(refs) < int(e) {
		return ""
	}
	return refs[e]
}

// MarshalJSON encodes type in to JSON.
func (e ElevationRefType) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

// UnmarshalJSON decodes type from JSON.
func (e *ElevationRefType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	return e.FromString(s)
}
