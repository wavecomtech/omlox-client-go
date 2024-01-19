// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"testing"
)

var providersJSONTestCases = []struct {
	name     string
	provider LocationProvider
	json     []byte
}{
	{
		name: "required",
		provider: LocationProvider{
			ID:   "ac:23:3f:ac:a3:87",
			Type: LocationProviderTypeUwb,
		},
		json: []byte(`{"id":"ac:23:3f:ac:a3:87","type":"uwb"}`),
	},
	{
		name: "fully-populated",
		provider: LocationProvider{
			ID:               "ac:23:3f:ac:a3:87",
			Type:             LocationProviderTypeIbeacon,
			Name:             "Minew Tag",
			Sensors:          map[string]any{"temp": float64(65.3)},
			FenceTimeout:     NewDuration(800),
			ExitTolerance:    1.3,
			ToleranceTimeout: NewDuration(Inf),
			ExitDelay:        NewDuration(Inf),
			Properties:       json.RawMessage(`{"org.wavecom.whereis":{"eid":"MINEW355"}}`),
		},
		json: []byte(`{"id":"ac:23:3f:ac:a3:87","type":"ibeacon","name":"Minew Tag","sensors":{"temp":65.3},"fence_timeout":800,"exit_tolerance":1.3,"tolerance_timeout":-1,"exit_delay":-1,"properties":{"org.wavecom.whereis":{"eid":"MINEW355"}}}`),
	},
}

func TestProviderMarshal(t *testing.T) {
	for _, tc := range providersJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONMarshalOK(t, tc.provider, tc.json)
		})
	}
}

func TestProviderUnmarshal(t *testing.T) {
	for _, tc := range providersJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONUnmarshalOK(t, tc.json, tc.provider)
		})
	}
}
