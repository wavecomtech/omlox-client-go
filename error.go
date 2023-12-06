// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"log/slog"
)

// Error is the error returned when Omlox Hub responds with an HTTP status
// code outside of the 200 - 399 range.  If a request fails due to a
// network error, a different error message will be returned.
type Error struct {
	// A text representation of the error type (required).
	Type string `json:"type"`

	// HTTP status code (required).
	Code int `json:"code"`

	// A human readable error message which may give a hint to what went wrong (optional).
	Message string `json:"message"`
}

// isResponseError determines if this is a response error based on the response
// status code. If it is determined to be an error, the function consumes the
// response body without closing it and parses the underlying error messages.
func isResponseError(r *http.Response) error {
	// 200 to 399 are non-error status codes
	if r.StatusCode >= 200 && r.StatusCode <= 399 {
		return nil
	}

	// read the entire response first so that we can return it as a raw error
	// in case in cannot be parsed
	responseBody, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("received an error response from omlox but could not read its body: %s", err.Error())
	}

	var responseError Error
	if err := json.Unmarshal(responseBody, &responseError); err != nil {
		// return the raw response body
		return errors.New(string(responseBody))
	}

	return &responseError
}

func (err Error) Error() string {
	return fmt.Sprintf("%s (code %d): %s", err.Type, err.Code, err.Message)
}

// LogValue implements [slog.LogValuer] to convert itself into a Value for logging.
func (err Error) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("type", err.Type),
		slog.Int("code", err.Code),
		slog.String("msg", err.Message),
	)
}
