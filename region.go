// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"errors"

	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
)

type Region struct {
	geojson.Object
}

func NewRegionPoint(point geometry.Point) *Region {
	return &Region{geojson.NewPoint(point)}
}

func NewRegionPolygon(poly *geometry.Poly) *Region {
	return &Region{geojson.NewPolygon(poly)}
}

func (r Region) MarshalJSON() ([]byte, error) {
	return r.Object.MarshalJSON()
}

func (r *Region) UnmarshalJSON(data []byte) error {
	o, err := geojson.Parse(string(data), geojson.DefaultParseOptions)
	if err != nil {
		return err
	}

	switch o.(type) {
	case *geojson.Point, *geojson.Polygon:
	default:
		return errors.New("region must be a geojson point or polygon")
	}

	*r = Region{
		Object: o,
	}

	return nil
}

func (r Region) Equal(u Region) bool {
	switch objectR := r.Object.(type) {
	case *geojson.Point:
		if objectU, ok := u.Object.(*geojson.Point); ok {
			return Point{*objectR}.Equal(Point{*objectU})
		}
	case *geojson.Polygon:
		if objectU, ok := u.Object.(*geojson.Polygon); ok {
			return Polygon{*objectR}.Equal(Polygon{*objectU})
		}
	}

	return false
}
