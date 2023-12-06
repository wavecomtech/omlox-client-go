// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"testing"
)

var wrapperObjJSONTestCases = []struct {
	name  string
	wrObj WrapperObject
	json  []byte
}{
	{
		name: "subcription",
		wrObj: WrapperObject{
			Event:  EventSubscribe,
			Topic:  TopicLocationUpdates,
			Params: Parameters{"crs": "EPSG:4326"},
		},
		json: []byte(`{"event":"subscribe","topic":"location_updates","params":{"crs":"EPSG:4326"}}`),
	},
	{
		name: "location-update",
		wrObj: WrapperObject{
			Event:          EventMsg,
			Topic:          TopicLocationUpdates,
			SubscriptionID: 123,
			Payload:        []json.RawMessage{json.RawMessage(`{"position":{"type":"Point","coordinates":[5,4]},"source":"fdb6df62-bce8-6c23-e342-80bd5c938774","provider_type":"uwb","provider_id":"77:4f:34:69:27:40","timestamp_generated":"2019-09-02T22:02:24.355Z","timestamp_sent":"2019-09-02T22:02:24.355Z"}`)},
		},
		json: []byte(`{"event":"message","topic":"location_updates","subscription_id":123,"payload":[{"position":{"type":"Point","coordinates":[5,4]},"source":"fdb6df62-bce8-6c23-e342-80bd5c938774","provider_type":"uwb","provider_id":"77:4f:34:69:27:40","timestamp_generated":"2019-09-02T22:02:24.355Z","timestamp_sent":"2019-09-02T22:02:24.355Z"}]}`),
	},
}

func TestWrapperObjectMarshal(t *testing.T) {
	for _, tc := range wrapperObjJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONMarshalOK(t, tc.wrObj, tc.json)
		})
	}
}

func TestWrapperObjectUnmarshal(t *testing.T) {
	for _, tc := range wrapperObjJSONTestCases {
		t.Run(tc.name, func(t *testing.T) {
			JSONUnmarshalOK(t, tc.json, tc.wrObj)
		})
	}
}
