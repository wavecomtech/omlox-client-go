// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Fence defines model for Fence.
//
//easyjson:json
type Fence struct {
	// ID must be a unique identifier (e.g. a UUID). When creating a fence, a unique id
	// will be generated if it is not provided.
	ID uuid.UUID `json:"id"`

	// Region is a GeoJSON Point or Polygon geometry defining the region of the geofence.
	Region *Region `json:"region"`

	// Radius defines a circular region around a point in meters when the region is a point.
	// The radius value is ignored for polygonal regions.
	Radius float64 `json:"radius,omitempty"`

	// Extrusion is the extrusion to be applied to the geometry in meters.
	// Must be a positive number.
	Extrusion float64 `json:"extrusion,omitempty"`

	// Floor is the canonical representation of the floor level, where floor 0 is the
	// ground floor.
	Floor float64 `json:"floor,omitempty"`

	// ForeignID represents a foreign unique identifier for the origin of a fence event,
	// such as an UWB zone id or an iBeacon identifier. If a foreign_id is set, the foreign
	// id MUST be resolved by the omlox™ hub.
	ForeignID string `json:"foreign_id,omitempty"`

	// Name is a textual representation of the fence.
	Name string `json:"name,omitempty"`

	// Timeout is the timeout in milliseconds after which a location should expire and
	// trigger a fence exit event (if no more location updates are sent). Must be a positive
	// number or -1 in case of an infinite timeout. If not set, or set to null, it will
	// default to infinite.
	Timeout Duration `json:"timeout,omitempty"`

	// ExitTolerance is the distance tolerance to a fence in meters before an exit event is
	// triggered. Useful for locations nearby or on the border of a fence to avoid fluctuating
	// fence entry / exit events. For example, a location which was previously inside a fence
	// will remain within the fence when its distance to the fence is less than or equal to
	// the given tolerance. Must be a positive number. If not set, or null, the exit_tolerance
	// will default to 0.
	ExitTolerance float64 `json:"exit_tolerance,omitempty"`

	// ToleranceTimeout is the timeout in milliseconds after which a location outside of a
	// fence, but still within exit_tolerance distance to that fence, should trigger a fence
	// exit event. For example, assume an exit_tolerance of 1m: A location previously within
	// the fence is now located 50cm outside of the fence and remains within the given
	// exit_tolerance distance to that fence. An exit event is triggered after tolerance_timeout
	// when the location remains outside of the fence. If not set, or null, the timeout will
	// be equal to the fence timeout. If tolerance_timeout is greater than the fence timeout
	// the tolerance_timeout will be reduced to the fence timeout. The provided number must be
	// positive or -1 in case of an infinite tolerance_timeout. If not set, or null,
	// tolerance_timeout will default to 0.
	ToleranceTimeout Duration `json:"tolerance_timeout,omitempty"`

	// ExitDelay is the delay in milliseconds in which an imminent exit event should wait for
	// another location update. This is relevant for fast rate position updates with quick
	// moving objects. For example, an RTLS provider may batch location updates into groups,
	// resulting in distances being temporarily outdated and premature events between quick
	// moving objects. The provided number must be positive or -1 in case of an infinite
	// exit_delay. If not set, or null, exit_delay will default to 0.
	ExitDelay Duration `json:"exit_delay,omitempty"`

	// Crs is a projection identifier defining the projection of the fence region coordinates.
	// The crs MUST be either a valid EPSG identifier (https://epsg.io) or 'local' if the fence
	// region is provided as relative coordinates of a zone. If crs is not present, 'EPSG:4326'
	// ("GPS") MUST be assumed as the default.
	Crs string `json:"crs,omitempty"`

	// ZoneId is the zone id related to the fence's region. This field MUST be present
	// in case crs projection is set to 'local'.
	ZoneID string `json:"zone_id,omitempty"`

	// ElevationRef is an elevation reference hint for the position's z component.
	// If present, it MUST be either 'floor' or 'wgs84'. If set to 'floor' the z component
	// MUST be assumed to be relative to the floor level. If set to 'wgs84' the z component
	// MUST be treated as WGS84 ellipsoidal height. For the majority of applications an accurate
	// geographic height may not be available. Therefore elevation_ref MUST be assumed 'floor'
	// by default if this property is not present.
	ElevationRef *ElevationRefType `json:"elevation_ref,omitempty"`

	// Properties contains any additional application or vendor specific properties.
	// An application implementing this object is not required to interpret any of the
	// custom properties, but it MUST preserve the properties if set.
	Properties json.RawMessage `json:"properties,omitempty"`
}
