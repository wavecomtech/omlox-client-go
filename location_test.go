// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
)

func mustParseTime(s string) *time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return &t
}

func opt[T any](v T) *T {
	return &v
}

var locationJSONTestCases = []struct {
	name     string
	location Location
	json     []byte
}{
	{
		name: "required",
		location: Location{
			Position:     Point{*geojson.NewPoint(geometry.Point{X: 7.815694, Y: 48.13021599999995})},
			Source:       "f4c05a2b-afd3-41a0-88e2-46f69bdb192e",
			ProviderType: LocationProviderTypeIbeacon,
			ProviderID:   "ac:23:3f:af:f3:90",
		},
		json: []byte(`{"position":{"type":"Point","coordinates":[7.815694,48.13021599999995]},"source":"f4c05a2b-afd3-41a0-88e2-46f69bdb192e","provider_type":"ibeacon","provider_id":"ac:23:3f:af:f3:90"}`),
	},
	{
		name: "fully-populated",
		location: Location{
			Position:     Point{*geojson.NewPointZ(geometry.Point{X: 7.815694, Y: 48.13021599999995}, 1.2)},
			Source:       "f4c05a2b-afd3-41a0-88e2-46f69bdb192e",
			ProviderType: LocationProviderTypeIbeacon,
			ProviderID:   "ac:23:3f:af:f3:90",
			Trackables: []uuid.UUID{
				uuid.MustParse("9d3b2ee3-791f-444d-a0f2-caf52820f561"),
				uuid.MustParse("a5865271-2e84-40d0-8f8f-e6f7ea15d103"),
			},
			TimestampGenerated: mustParseTime("2023-10-17T11:14:37.206Z"),
			TimestampSent:      mustParseTime("2023-10-17T11:14:37.213Z"),
			Crs:                "local",
			Associated:         true,
			Floor:              1.2,
			TrueHeading:        opt(-1.0),
			MagneticHeading:    opt(1.234),
			HeadingAccuracy:    opt(1.12),
			ElevationRef:       opt(ElevationRefTypeWgs84),
			Speed:              opt(0.814870001487674),
			Course:             opt(104.76595042882053),
			Properties:         json.RawMessage(`{"org.wavecom.temp":24.3}`),
		},
		json: []byte(`{"position":{"type":"Point","coordinates":[7.815694,48.13021599999995,1.2]},"source":"f4c05a2b-afd3-41a0-88e2-46f69bdb192e","provider_type":"ibeacon","provider_id":"ac:23:3f:af:f3:90","trackables":["9d3b2ee3-791f-444d-a0f2-caf52820f561","a5865271-2e84-40d0-8f8f-e6f7ea15d103"],"timestamp_generated":"2023-10-17T11:14:37.206Z","timestamp_sent":"2023-10-17T11:14:37.213Z","crs":"local","associated":true,"floor":1.2,"true_heading":-1,"magnetic_heading":1.234,"heading_accuracy":1.12,"elevation_ref":"wgs84","speed":0.814870001487674,"course":104.76595042882053,"properties":{"org.wavecom.temp":24.3}}`),
	},
}

func TestLocationMarshal(t *testing.T) {
	for _, tc := range locationJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONMarshalOK(t, tc.location, tc.json)
		})
	}
}

func TestLocationUnmarshal(t *testing.T) {
	for _, tc := range locationJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONUnmarshalOK(t, tc.json, tc.location)
		})
	}
}
