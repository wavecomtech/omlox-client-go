// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"

	"github.com/google/uuid"
)

// TrackableMotion defines model for TrackableMotion.
// It describes a trackable's position and shape update.
//
//easyjson:json
type TrackableMotion struct {
	// The trackable UUID for which this motion data was generated.
	ID uuid.UUID `json:"id"`

	// The name of the trackable.
	Name string `json:"name,omitempty"`

	// GeoJson Polygon geometry. Important: A Polygon object MUST be interpreted according to a coordinate reference system (crs).
	// Therefore, coordinates MUST match the CRS which MUST be either a valid EPSG identifier (https://epsg.io) or
	// 'local' if it is provided as a relative coordinate.
	// The ordering of components is x,y,z or longitude,latitude,elevation respectively as according to the GeoJson specification.
	Geometry *Polygon `json:"geometry,omitempty"`

	// The extrusion to be applied to the geometry in meters.
	// Must be a positive number.
	Extrusion float64 `json:"extrusion,omitempty"`

	// The location for the trackable which triggered this motion update.
	Location Location `json:"location"`

	// A copy of application or vendor specific properties from the Trackable for which this TrackableMotion was generated.
	// An application implementing this object is not required to interpret any of the custom properties,
	// but it MUST preserve the properties if set.
	// Properties MUST be copied from the related Trackable to this TrackableMotion object.
	Properties json.RawMessage `json:"properties,omitempty"`
}
