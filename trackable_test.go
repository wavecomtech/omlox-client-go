// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/tidwall/geojson/geometry"
)

var trackablesJSONTestCases = []struct {
	name      string
	trackable Trackable
	json      []byte
}{
	{
		name: "required",
		trackable: Trackable{
			ID:   uuid.MustParse("9b59961e-2a6a-4712-86e7-aba5a3e8be1f"),
			Type: TrackableTypeOmlox,
		},
		json: []byte(`{"id":"9b59961e-2a6a-4712-86e7-aba5a3e8be1f","type":"omlox"}`),
	},
	{
		name: "fully-populated",
		trackable: Trackable{
			ID:   uuid.MustParse("9b59961e-2a6a-4712-86e7-aba5a3e8be1f"),
			Type: TrackableTypeVirtual,
			Name: "Container",
			Geometry: NewPolygon(geometry.NewPoly([]geometry.Point{
				{X: 7.815694, Y: 48.13021599999995},
				{X: 7.815724999999997, Y: 48.13031},
				{X: 7.816582, Y: 48.13018799999995},
				{X: 7.816551, Y: 48.13009399999996},
				{X: 7.815694, Y: 48.13021599999995},
			}, nil, geometry.DefaultIndexOptions)),
			Extrusion:         1.22,
			LocationProviders: []string{"ac:23:3f:ac:a3:55"},
			FenceTimeout:      NewDuration(Inf),
			ExitTolerance:     1,
			ToleranceTimeout:  NewDuration(Inf),
			ExitDelay:         NewDuration(2),
			Properties:        json.RawMessage(`{"org.wavecom.whereis":{"eid":"CTR0008"}}`),
		},
		json: []byte(`{"id":"9b59961e-2a6a-4712-86e7-aba5a3e8be1f","type":"virtual","name":"Container","geometry":{"type":"Polygon","coordinates":[[[7.815694,48.13021599999995],[7.815724999999997,48.13031],[7.816582,48.13018799999995],[7.816551,48.13009399999996],[7.815694,48.13021599999995]]]},"extrusion":1.22,"location_providers":["ac:23:3f:ac:a3:55"],"fence_timeout":-1,"exit_tolerance":1,"tolerance_timeout":-1,"exit_delay":2,"properties":{"org.wavecom.whereis":{"eid":"CTR0008"}}}`),
	},
}

func TestTrackableMarshal(t *testing.T) {
	for _, tc := range trackablesJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONMarshalOK(t, tc.trackable, tc.json)
		})
	}
}

func TestTrackableUnmarshal(t *testing.T) {
	for _, tc := range trackablesJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONUnmarshalOK(t, tc.json, tc.trackable)
		})
	}
}
