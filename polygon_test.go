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

var polygonJSONTestCases = []struct {
	poly *Polygon
	json []byte
}{
	{
		poly: NewPolygon(geometry.NewPoly([]geometry.Point{
			{X: 7.815694, Y: 48.13021599999995},
			{X: 7.815724999999997, Y: 48.13031},
			{X: 7.816582, Y: 48.13018799999995},
			{X: 7.816551, Y: 48.13009399999996},
			{X: 7.815694, Y: 48.13021599999995},
		}, nil, geometry.DefaultIndexOptions)),
		json: []byte(`{"type":"Polygon","coordinates":[[[7.815694,48.13021599999995],[7.815724999999997,48.13031],[7.816582,48.13018799999995],[7.816551,48.13009399999996],[7.815694,48.13021599999995]]]}`),
	},
}

func TestPolygonMarshal(t *testing.T) {
	for _, tc := range polygonJSONTestCases {
		output, err := json.Marshal(tc.poly)
		if err != nil {
			t.Fatal(err)
		}

		opts := jsondiff.DefaultConsoleOptions()
		if r, diff := jsondiff.Compare(tc.json, output, &opts); r != jsondiff.FullMatch {
			t.Errorf("%s", diff)
		}
	}
}

func TestPolygonUnmarshal(t *testing.T) {
	for _, tc := range polygonJSONTestCases {
		var poly Polygon
		if err := json.Unmarshal(tc.json, &poly); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(tc.poly, &poly); diff != "" {
			t.Errorf("Polygon() mismatch (-want +got):\n%s", diff)
		}
	}

}

var polygonEqualTestCases = []struct {
	name     string
	p        *Polygon
	u        *Polygon
	expected bool
}{
	{
		name: "identical-polygons",
		p: NewPolygon(geometry.NewPoly([]geometry.Point{
			{X: 0.0, Y: 0.0},
			{X: 1.0, Y: 0.0},
			{X: 1.0, Y: 1.0},
			{X: 0.0, Y: 1.0},
			{X: 0.0, Y: 0.0},
		}, nil, geometry.DefaultIndexOptions)),
		u: NewPolygon(geometry.NewPoly([]geometry.Point{
			{X: 0.0, Y: 0.0},
			{X: 1.0, Y: 0.0},
			{X: 1.0, Y: 1.0},
			{X: 0.0, Y: 1.0},
			{X: 0.0, Y: 0.0},
		}, nil, geometry.DefaultIndexOptions)),
		expected: true,
	},
	{
		name: "identical-polygons-rotated",
		p: NewPolygon(geometry.NewPoly([]geometry.Point{
			{X: 0.0, Y: 0.0},
			{X: 1.0, Y: 0.0},
			{X: 1.0, Y: 1.0},
			{X: 0.0, Y: 1.0},
			{X: 0.0, Y: 0.0},
		}, nil, geometry.DefaultIndexOptions)),
		u: NewPolygon(geometry.NewPoly([]geometry.Point{
			{X: 1.0, Y: 0.0},
			{X: 1.0, Y: 1.0},
			{X: 0.0, Y: 1.0},
			{X: 0.0, Y: 0.0},
			{X: 1.0, Y: 0.0},
		}, nil, geometry.DefaultIndexOptions)),
		expected: true,
	},
	{
		name: "different-polygons",
		p: NewPolygon(geometry.NewPoly([]geometry.Point{
			{X: 0.0, Y: 0.0},
			{X: 1.0, Y: 0.0},
			{X: 1.0, Y: 1.0},
			{X: 0.0, Y: 1.0},
			{X: 0.0, Y: 0.0},
		}, nil, geometry.DefaultIndexOptions)),
		u: NewPolygon(geometry.NewPoly([]geometry.Point{
			{X: 1.0, Y: 1.0},
			{X: 2.0, Y: 1.0},
			{X: 2.0, Y: 2.0},
			{X: 1.0, Y: 2.0},
			{X: 1.0, Y: 1.0},
		}, nil, geometry.DefaultIndexOptions)),
		expected: false,
	},
	{
		name: "polygon-within-polygon",
		p: NewPolygon(geometry.NewPoly([]geometry.Point{
			{X: 0.0, Y: 0.0},
			{X: 1.0, Y: 0.0},
			{X: 1.0, Y: 1.0},
			{X: 0.0, Y: 1.0},
			{X: 0.0, Y: 0.0},
		}, nil, geometry.DefaultIndexOptions)),
		u: NewPolygon(geometry.NewPoly([]geometry.Point{
			{X: -1.0, Y: -1.0},
			{X: 2.0, Y: -1.0},
			{X: 2.0, Y: 2.0},
			{X: -1.0, Y: 2.0},
			{X: -1.0, Y: -1.0},
		}, nil, geometry.DefaultIndexOptions)),
		expected: false,
	},
}

func TestPolygonEqual(t *testing.T) {
	for _, tc := range polygonEqualTestCases {
		t.Run(tc.name, func(t *testing.T) {
			if result := tc.p.Equal(*tc.u); result != tc.expected {
				t.Errorf("Equal() mismatch want: %v, got: %v", tc.expected, result)
			}
		})
	}
}
