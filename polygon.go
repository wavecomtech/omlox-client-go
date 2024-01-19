// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"errors"
	"strings"

	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
)

type Polygon struct {
	geojson.Polygon
}

func NewPolygon(poly *geometry.Poly) *Polygon {
	return &Polygon{*geojson.NewPolygon(poly)}
}

func (p Polygon) MarshalJSON() ([]byte, error) {
	return p.Polygon.MarshalJSON()
}

func (p *Polygon) UnmarshalJSON(data []byte) error {
	o, err := geojson.Parse(string(data), geojson.DefaultParseOptions)
	if err != nil {
		return err
	}

	poly, ok := o.(*geojson.Polygon)
	if !ok {
		return errors.New("not a valid geojson polygon")
	}

	*p = Polygon{
		Polygon: *poly,
	}

	return nil
}

func (p Polygon) Equal(u Polygon) bool {
	if p.NumPoints() != u.NumPoints() {
		return false
	}

	if strings.Compare(p.JSON(), u.JSON()) != 0 {
		return false
	}

	return true
}
