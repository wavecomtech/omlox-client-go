// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package output

import (
	"fmt"
	"io"
)

type Format string

const (
	Table Format = "table"
	JSON  Format = "json"
)

// Formats returns a list of the string representation of the supported formats
func Formats() []string {
	return []string{Table.String(), JSON.String()}
}

// FormatsWithDesc returns a list of the string representation of the supported formats
// including a description
func FormatsWithDesc() map[string]string {
	return map[string]string{
		Table.String(): "Output result in human-readable format",
		JSON.String():  "Output result in JSON format",
	}
}

// ErrInvalidFormatType is returned when an unsupported format type is used
var ErrInvalidFormatType = fmt.Errorf("invalid format type")

// String returns the string representation of the Format
func (o Format) String() string {
	return string(o)
}

// Write the output in the given format to the io.Writer. Unsupported formats
// will return an error
func (o Format) Write(out io.Writer, w Writer) error {
	switch o {
	case Table:
		return w.WriteTable(out)
	case JSON:
		return w.WriteJSON(out)
	}
	return ErrInvalidFormatType
}

// ParseFormat takes a raw string and returns the matching Format.
// If the format does not exists, ErrInvalidFormatType is returned
func ParseFormat(s string) (out Format, err error) {
	switch s {
	case Table.String():
		out, err = Table, nil
	case JSON.String():
		out, err = JSON, nil
	default:
		out, err = "", ErrInvalidFormatType
	}
	return
}

// Writer is an interface that any type can implement to write supported formats
type Writer interface {
	// WriteTable will write tabular output into the given io.Writer, returning
	// an error if any occur
	WriteTable(out io.Writer) error
	// WriteJSON will write JSON formatted output into the given io.Writer,
	// returning an error if any occur
	WriteJSON(out io.Writer) error
}
