// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"errors"

	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
)

type Point struct {
	geojson.Point
}

func NewPoint(point geometry.Point) *Point {
	return &Point{*geojson.NewPoint(point)}
}

func NewPointZ(point geometry.Point, z float64) *Point {
	return &Point{*geojson.NewPointZ(point, z)}
}

func (p Point) MarshalJSON() ([]byte, error) {
	return p.Point.MarshalJSON()
}

func (p *Point) UnmarshalJSON(data []byte) error {
	o, err := geojson.Parse(string(data), geojson.DefaultParseOptions)
	if err != nil {
		return err
	}

	point, ok := o.(*geojson.Point)
	if !ok {
		return errors.New("not a valid geojson point")
	}

	*p = Point{
		Point: *point,
	}

	return nil
}

func (p Point) Equal(u Point) bool {
	return p.WithinPoint(u.Base())
}
