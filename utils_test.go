// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nsf/jsondiff"
)

// JSONMarshalOK utility function to check that structs are marshalled correctly to json.
func JSONMarshalOK[T any](t *testing.T, input T, expected []byte) {
	t.Helper()

	o, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}

	opts := jsondiff.DefaultConsoleOptions()
	if r, diff := jsondiff.Compare(expected, o, &opts); r != jsondiff.FullMatch {
		t.Fatalf("%s", diff)
	}
}

func JSONUnmarshalOK[T any](t *testing.T, input []byte, expected T) {
	var (
		o   T
		err error
	)

	tp := reflect.ValueOf(expected)
	if tp.Kind() == reflect.Ptr {
		t.Fatal("unsupported pointer value. use a value instead.")
	}

	if err = json.Unmarshal(input, &o); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, o); diff != "" {
		t.Errorf("%s mismatch (-want +got):\n%s", tp.String(), diff)
	}
}
