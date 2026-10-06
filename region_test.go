// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nsf/jsondiff"
	"github.com/tidwall/geojson/geometry"
)

var regionJSONTestCases = []struct {
	reg  *Region
	json []byte
}{
	{
		reg: NewRegionPolygon(geometry.NewPoly([]geometry.Point{
			{X: 7.815694, Y: 48.13021599999995},
			{X: 7.815724999999997, Y: 48.13031},
			{X: 7.816582, Y: 48.13018799999995},
			{X: 7.816551, Y: 48.13009399999996},
			{X: 7.815694, Y: 48.13021599999995},
		}, nil, geometry.DefaultIndexOptions)),
		json: []byte(`{"type":"Polygon","coordinates":[[[7.815694,48.13021599999995],[7.815724999999997,48.13031],[7.816582,48.13018799999995],[7.816551,48.13009399999996],[7.815694,48.13021599999995]]]}`),
	},
	{
		reg:  NewRegionPoint(geometry.Point{X: 7.815694, Y: 48.13021599999995}),
		json: []byte(`{"type":"Point","coordinates":[7.815694,48.13021599999995]}`),
	},
}

func TestRegionMarshal(t *testing.T) {
	for _, tc := range regionJSONTestCases {
		output, err := json.Marshal(tc.reg)
		if err != nil {
			t.Fatal(err)
		}

		opts := jsondiff.DefaultConsoleOptions()
		if r, diff := jsondiff.Compare(tc.json, output, &opts); r != jsondiff.FullMatch {
			t.Errorf("%s", diff)
		}
	}
}

func TestRegionUnmarshal(t *testing.T) {
	for _, tc := range regionJSONTestCases {
		var reg Region
		if err := json.Unmarshal(tc.json, &reg); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(tc.reg, &reg); diff != "" {
			t.Errorf("Region() mismatch (-want +got):\n%s", diff)
		}
	}
}

var regionEqualTestCases = []struct {
	name     string
	r        *Region
	u        *Region
	expected bool
}{
	{
		name: "different-region-objects",
		r:    NewRegionPoint(geometry.Point{X: 7.815694, Y: 48.13021599999995}),
		u: NewRegionPolygon(geometry.NewPoly([]geometry.Point{
			{X: 0.0, Y: 0.0},
			{X: 1.0, Y: 0.0},
			{X: 1.0, Y: 1.0},
			{X: 0.0, Y: 1.0},
			{X: 0.0, Y: 0.0},
		}, nil, geometry.DefaultIndexOptions)),
		expected: false,
	},
}

func TestRegionEqual(t *testing.T) {
	for _, tc := range regionEqualTestCases {
		t.Run(tc.name, func(t *testing.T) {
			if result := tc.r.Equal(*tc.u); result != tc.expected {
				t.Errorf("Equal() mismatch want: %v, got: %v", tc.expected, result)
			}
		})
	}
}
