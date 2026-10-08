// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/tidwall/geojson"
	"github.com/tidwall/geojson/geometry"
)

var trackableMotionsJSONTestCases = []struct {
	name            string
	trackableMotion TrackableMotion
	json            []byte
}{
	{
		name: "required",
		trackableMotion: TrackableMotion{
			ID: uuid.MustParse("9b59961e-2a6a-4712-86e7-aba5a3e8be1f"),
			Location: Location{
				Position:     Point{*geojson.NewPoint(geometry.Point{X: 7.815694, Y: 48.13021599999995})},
				Source:       "f4c05a2b-afd3-41a0-88e2-46f69bdb192e",
				ProviderType: LocationProviderTypeIbeacon,
				ProviderID:   "ac:23:3f:af:f3:90",
			},
		},
		json: []byte(`{"id":"9b59961e-2a6a-4712-86e7-aba5a3e8be1f","location":{"position":{"type":"Point","coordinates":[7.815694,48.13021599999995]},"source":"f4c05a2b-afd3-41a0-88e2-46f69bdb192e","provider_type":"ibeacon","provider_id":"ac:23:3f:af:f3:90"}}`),
	},
	{
		name: "fully-populated",
		trackableMotion: TrackableMotion{
			ID:   uuid.MustParse("9b59961e-2a6a-4712-86e7-aba5a3e8be1f"),
			Name: "Container",
			Geometry: NewPolygon(geometry.NewPoly([]geometry.Point{
				{X: 7.815694, Y: 48.13021599999995},
				{X: 7.815724999999997, Y: 48.13031},
				{X: 7.816582, Y: 48.13018799999995},
				{X: 7.816551, Y: 48.13009399999996},
				{X: 7.815694, Y: 48.13021599999995},
			}, nil, geometry.DefaultIndexOptions)),
			Extrusion: 1.22,
			Location: Location{
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
			Properties: json.RawMessage(`{"org.wavecom.whereis":{"eid":"CTR0008"}}`),
		},
		json: []byte(`{"id":"9b59961e-2a6a-4712-86e7-aba5a3e8be1f","name":"Container","geometry":{"type":"Polygon","coordinates":[[[7.815694,48.13021599999995],[7.815724999999997,48.13031],[7.816582,48.13018799999995],[7.816551,48.13009399999996],[7.815694,48.13021599999995]]]},"extrusion":1.22,"location":{"position":{"type":"Point","coordinates":[7.815694,48.13021599999995,1.2]},"source":"f4c05a2b-afd3-41a0-88e2-46f69bdb192e","provider_type":"ibeacon","provider_id":"ac:23:3f:af:f3:90","trackables":["9d3b2ee3-791f-444d-a0f2-caf52820f561","a5865271-2e84-40d0-8f8f-e6f7ea15d103"],"timestamp_generated":"2023-10-17T11:14:37.206Z","timestamp_sent":"2023-10-17T11:14:37.213Z","crs":"local","associated":true,"floor":1.2,"true_heading":-1,"magnetic_heading":1.234,"heading_accuracy":1.12,"elevation_ref":"wgs84","speed":0.814870001487674,"course":104.76595042882053,"properties":{"org.wavecom.temp":24.3}},"properties":{"org.wavecom.whereis":{"eid":"CTR0008"}}}`),
	},
}

func TestTrackableMotionMarshal(t *testing.T) {
	for _, tc := range trackableMotionsJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONMarshalOK(t, tc.trackableMotion, tc.json)
		})
	}
}

func TestTrackableMotionUnmarshal(t *testing.T) {
	for _, tc := range trackableMotionsJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONUnmarshalOK(t, tc.json, tc.trackableMotion)
		})
	}
}
