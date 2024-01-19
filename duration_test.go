// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestDurationMarshal(t *testing.T) {
	cases := []struct {
		val Duration
		out []byte
	}{
		{Duration{}, []byte("null")},
		{Duration{dur: 1}, []byte("null")},
		{Duration{dur: -1}, []byte("null")},
		{Duration{dur: 0, defined: true}, []byte("0")},
		{Duration{dur: 5, defined: true}, []byte("5")},
		{Duration{dur: -1, defined: true}, []byte("-1")},
		{Duration{dur: -5, defined: true}, []byte("-1")},
	}

	for _, test := range cases {
		t.Run(fmt.Sprintf("Duration(dur=%d,defined=%v)=json(%q)", test.val.dur, test.val.defined, test.out), func(t *testing.T) {
			data, err := json.Marshal(test.val)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(data, test.out) {
				t.Errorf("mismatch! wanted: '%s' and got '%s'", test.out, data)
			}
		})
	}
}

func TestDurationUnmarshal(t *testing.T) {
	cases := []struct {
		in  []byte
		val Duration
	}{
		{[]byte("null"), Duration{}},
		{[]byte("-1"), Duration{dur: -1, defined: true}},
		{[]byte("0"), Duration{dur: 0, defined: true}},
		{[]byte("1"), Duration{dur: 1, defined: true}},
		{[]byte("99999"), Duration{dur: 99999, defined: true}},
		{[]byte("-99999"), Duration{dur: -1, defined: true}},
	}

	for _, test := range cases {
		t.Run(fmt.Sprintf("json(%q)=Duration(dur=%d,defined=%v)", test.in, test.val.dur, test.val.defined), func(t *testing.T) {
			var dur Duration
			if err := json.Unmarshal(test.in, &dur); err != nil {
				t.Fatal(err)
			}

			if dur.dur != test.val.dur || dur.defined != test.val.defined {
				t.Errorf("mismatch! wanted: %v and got %v", reprDuration(test.val), reprDuration(dur))
			}
		})
	}
}

func TestDurationEqual(t *testing.T) {
	cases := []struct {
		x, y  Duration
		equal bool
	}{
		{NewDuration(Inf), NewDuration(Inf), true},
		{NewDuration(Inf), NewDuration(-99), true},
		{NewDuration(90), NewDuration(Inf), false},
		{NewDuration(300), NewDuration(300), true},
		{NewDuration(0), NewDuration(500), false},
		{Duration{defined: true}, Duration{}, false},
		{Duration{defined: true, dur: 6}, Duration{dur: 6}, false},
	}

	for i, tc := range cases {
		if tc.x.Equal(tc.y) != tc.equal {
			t.Errorf("%d: %v == %v is %v and got %v", i, tc.x, tc.y, tc.equal, !tc.equal)
		}
	}
}

// reprDuration returns a representable Duration.
func reprDuration(milli Duration) string {
	return fmt.Sprintf("Duration(dur=%d,defined=%v)", milli.dur, milli.defined)
}
