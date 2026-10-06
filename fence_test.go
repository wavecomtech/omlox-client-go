// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/tidwall/geojson/geometry"
)

var fencesJSONTestCases = []struct {
	name  string
	fence Fence
	json  []byte
}{
	{
		name: "required",
		fence: Fence{
			ID: uuid.MustParse("497f6eca-6276-4993-bfeb-53cbbbba6f08"),
			Region: NewRegionPolygon(geometry.NewPoly([]geometry.Point{
				{X: 7.815694, Y: 48.13021599999995},
				{X: 7.815724999999997, Y: 48.13031},
				{X: 7.816582, Y: 48.13018799999995},
				{X: 7.816551, Y: 48.13009399999996},
				{X: 7.815694, Y: 48.13021599999995},
			}, nil, geometry.DefaultIndexOptions)),
		},
		json: []byte(`{"id":"497f6eca-6276-4993-bfeb-53cbbbba6f08","region":{"type":"Polygon","coordinates":[[[7.815694,48.13021599999995],[7.815724999999997,48.13031],[7.816582,48.13018799999995],[7.816551,48.13009399999996],[7.815694,48.13021599999995]]]}}`),
	},
	{
		name: "fully-populated",
		fence: Fence{
			ID:               uuid.MustParse("497f6eca-6276-4993-bfeb-53cbbbba6f08"),
			Region:           NewRegionPoint(geometry.Point{X: 7.815694, Y: 48.13021599999995}),
			Radius:           10,
			Extrusion:        1.22,
			Floor:            4,
			ForeignID:        "Beacon1C5: Floor 1",
			Name:             "Warehouse",
			Timeout:          NewDuration(5),
			ExitTolerance:    1,
			ToleranceTimeout: NewDuration(Inf),
			ExitDelay:        NewDuration(2),
			Crs:              "local",
			ZoneID:           "ecadd57B-7e4a-4636-b105-6b1c11bb75bf",
			ElevationRef:     opt(ElevationRefTypeWgs84),
			Properties:       json.RawMessage(`{"org.wavecom.whereis":{"site":"Wavecom"}}`),
		},
		json: []byte(`{"id":"497f6eca-6276-4993-bfeb-53cbbbba6f08","region":{"type":"Point","coordinates":[7.815694,48.13021599999995]},"radius":10,"extrusion":1.22,"floor":4,"foreign_id":"Beacon1C5: Floor 1","name":"Warehouse","timeout":5,"exit_tolerance":1,"tolerance_timeout":-1,"exit_delay":2,"crs":"local","zone_id":"ecadd57B-7e4a-4636-b105-6b1c11bb75bf","elevation_ref":"wgs84","properties":{"org.wavecom.whereis":{"site":"Wavecom"}}}`),
	},
}

func TestFenceMarshal(t *testing.T) {
	for _, tc := range fencesJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONMarshalOK(t, tc.fence, tc.json)
		})
	}
}

func TestFenceUnmarshal(t *testing.T) {
	for _, tc := range fencesJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONUnmarshalOK(t, tc.json, tc.fence)
		})
	}
}
