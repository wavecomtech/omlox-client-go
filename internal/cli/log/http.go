// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package log

import (
	"fmt"
	"net/http"
	"time"

	"log/slog"
)

type SlogerRoundTripper struct {
	Logger *slog.Logger
	Base   http.RoundTripper
}

var _ http.RoundTripper = (*SlogerRoundTripper)(nil)

func (l *SlogerRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	start := time.Now()
	res, err := l.Base.RoundTrip(r)
	dur := time.Since(start)

	l.Logger.LogAttrs(r.Context(), slog.LevelDebug, describeHttpRequest(r),
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
		slog.Int("status", res.StatusCode),
		slog.String("statusLabel", statusLabel(res.StatusCode)),
		slog.Duration("duration", dur),
	)

	return res, err
}

// describeHttpRequest returns a message describing the request .
func describeHttpRequest(r *http.Request) string {
	return fmt.Sprintf("Request %s %s", r.Method, r.URL.Path)
}

//nolint:gomnd
func statusLabel(status int) string {
	switch {
	case status >= 100 && status < 300:
		return "OK"
	case status >= 300 && status < 400:
		return "Redirect"
	case status >= 400 && status < 500:
		return "Client Error"
	case status >= 500:
		return "Server Error"
	default:
		return "Unknown"
	}
}
