// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"testing"
)

var errorsJSONTestCases = []struct {
	err  Error
	json []byte
}{
	{
		err: Error{
			Type:    "not found",
			Code:    404,
			Message: "Failed to get trackable with ID d27047bd-1b6b-4656-bb93-2326a4c900e1. Trackable does not exists.",
		},
		json: []byte(`{"type":"not found","code":404,"message":"Failed to get trackable with ID d27047bd-1b6b-4656-bb93-2326a4c900e1. Trackable does not exists."}`),
	},
}

func TestErrorMarshal(t *testing.T) {
	for _, tc := range errorsJSONTestCases {
		JSONMarshalOK(t, tc.err, tc.json)
	}
}

func TestErrorUnmarshal(t *testing.T) {
	for _, tc := range errorsJSONTestCases {
		JSONUnmarshalOK(t, tc.json, tc.err)
	}
}
