// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"testing"
	"time"
)

// WithReconnect predates the WithWS* options and is kept for existing callers,
// so it must keep enabling unlimited reconnection with the waits it was given.
func TestWithReconnect(t *testing.T) {
	tests := []struct {
		name    string
		min     time.Duration
		max     time.Duration
		wantErr bool
	}{
		{name: "valid", min: time.Second, max: 30 * time.Second},
		{name: "equal waits", min: time.Second, max: time.Second},
		{name: "zero min", min: 0, max: time.Second, wantErr: true},
		{name: "zero max", min: time.Second, max: 0, wantErr: true},
		{name: "min exceeds max", min: 2 * time.Second, max: time.Second, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := DefaultConfiguration()

			err := WithReconnect(tt.min, tt.max)(&c)
			if (err != nil) != tt.wantErr {
				t.Fatalf("WithReconnect() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if !c.WSAutoReconnect {
				t.Error("WSAutoReconnect = false, want true")
			}
			if c.WSMaxRetries != -1 {
				t.Errorf("WSMaxRetries = %d, want -1 (unlimited)", c.WSMaxRetries)
			}
			if c.WSMinRetryWait != tt.min || c.WSMaxRetryWait != tt.max {
				t.Errorf("retry wait = [%v, %v], want [%v, %v]", c.WSMinRetryWait, c.WSMaxRetryWait, tt.min, tt.max)
			}
		})
	}
}
