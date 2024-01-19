// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"time"

	easyjson "github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
)

const (
	// Inf is the constant value representing infite.
	Inf = -1

	maxDuration time.Duration = 1<<63 - 1
)

var _ easyjson.Optional = (*Duration)(nil)
var _ easyjson.MarshalerUnmarshaler = (*Duration)(nil)
var _ json.Marshaler = (*Duration)(nil)
var _ json.Unmarshaler = (*Duration)(nil)

// Duration is an Omlox type that provides infinite semantics for a time duration.
// Must be a positive number or -1 in case of an infinite duration.
// All negative durations are considered infinite.
type Duration struct {
	dur     int
	defined bool
}

// Create an infinite type of millisecond duration type.
func NewDuration(v int) Duration {
	if v <= Inf {
		return Duration{dur: Inf, defined: true}
	}

	return Duration{dur: v, defined: true}
}

// Inf return true on infinite value.
func (v Duration) Inf() bool {
	return v.dur <= Inf
}

// MarshalEasyJSON does JSON marshaling using easyjson interface.
func (v Duration) MarshalEasyJSON(w *jwriter.Writer) {
	if !v.defined {
		w.RawString("null")
		return
	}
	if v.dur <= Inf {
		w.Int(Inf)
	} else {
		w.Int(v.dur)
	}
}

// UnmarshalEasyJSON does JSON unmarshaling using easyjson interface.
func (v *Duration) UnmarshalEasyJSON(l *jlexer.Lexer) {
	if l.IsNull() {
		l.Skip()
		*v = Duration{}
	} else {
		*v = NewDuration(l.Int())
	}
}

// MarshalJSON implements a standard json marshaler interface.
// If the value is infinite, it will be ignored.
func (v Duration) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	v.MarshalEasyJSON(&w)
	return w.Buffer.BuildBytes(), w.Error
}

// UnmarshalJSON implements a standard json unmarshaler interface.
func (v *Duration) UnmarshalJSON(data []byte) error {
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)
	return l.Error()
}

// Duration returns a time.Duration.
// If its a infinite duration, it returns a maximum duration possible.
func (v *Duration) Duration() time.Duration {
	if v.Inf() {
		return maxDuration
	}

	return time.Duration(v.dur) * time.Millisecond
}

// IsDefined returns whether the value is defined.
// A function is required so that it can be used as [easyjson.Optional] interface.
func (v Duration) IsDefined() bool {
	return v.defined
}

// String implements a stringer interface using fmt.Sprint for the value.
func (v Duration) String() string {
	if v.Inf() {
		return "<inf>"
	}
	return v.Duration().String()
}

func (v Duration) Equal(y Duration) bool {
	if v.defined != y.defined {
		return false
	}

	if v.Inf() && y.Inf() {
		return true
	}

	if v.dur == y.dur {
		return true
	}

	return false
}
