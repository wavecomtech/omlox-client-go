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

var pointJSONTestCases = []struct {
	point *Point
	json  []byte
}{
	{
		point: NewPoint(geometry.Point{X: 7.815694, Y: 48.13021599999995}),
		json:  []byte(`{"type":"Point","coordinates":[7.815694,48.13021599999995]}`),
	},
}

func TestPointMarshal(t *testing.T) {
	for _, tc := range pointJSONTestCases {
		output, err := json.Marshal(tc.point)
		if err != nil {
			t.Fatal(err)
		}

		opts := jsondiff.DefaultConsoleOptions()
		if r, diff := jsondiff.Compare(tc.json, output, &opts); r != jsondiff.FullMatch {
			t.Errorf("%s", diff)
		}
	}
}

func TestPointUnmarshal(t *testing.T) {
	for _, tc := range pointJSONTestCases {
		var point Point
		if err := json.Unmarshal(tc.json, &point); err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(tc.point, &point); diff != "" {
			t.Errorf("Point() mismatch (-want +got):\n%s", diff)
		}
	}
}

var pointEqualTestCases = []struct {
	name     string
	p        *Point
	u        *Point
	expected bool
}{
	{
		name:     "equal-points",
		p:        NewPoint(geometry.Point{X: 7.815694, Y: 48.13021599999995}),
		u:        NewPoint(geometry.Point{X: 7.815694, Y: 48.13021599999995}),
		expected: true,
	},
	{
		name:     "different-points",
		p:        NewPoint(geometry.Point{X: 7.815694, Y: 48.13021599999995}),
		u:        NewPoint(geometry.Point{X: 7.815694, Y: 48.13021599999996}),
		expected: false,
	},
	{
		name:     "reverted-coordinates-points",
		p:        NewPoint(geometry.Point{X: 7.815694, Y: 48.13021599999995}),
		u:        NewPoint(geometry.Point{X: 48.13021599999995, Y: 7.815694}),
		expected: false,
	},
}

func TestPointEqual(t *testing.T) {
	for _, tc := range pointEqualTestCases {
		t.Run(tc.name, func(t *testing.T) {
			if result := tc.p.Equal(*tc.u); result != tc.expected {
				t.Errorf("Equal() mismatch want: %v, got: %v", tc.expected, result)
			}
		})
	}
}
