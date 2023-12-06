// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package resource

import (
	"bytes"
	"encoding/json"
	"io"
)

type Loader[T any] struct {
	Resources []T
}

const (
	tokenArrayStart = '['
)

// LoadJSON decode the provider reader stream in json format.
// The content can be an array or single object.
func (loader *Loader[T]) LoadJSON(r io.Reader) error {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}

	char, _, err := buf.ReadRune()
	if err != nil {
		return err
	}

	if err := buf.UnreadRune(); err != nil {
		return err
	}

	if char == tokenArrayStart {
		var resources []T
		if err := json.Unmarshal(buf.Bytes(), &resources); err != nil {
			return err
		}

		loader.Resources = append(loader.Resources, resources...)
		return nil
	}

	var resource T
	if err := json.Unmarshal(buf.Bytes(), &resource); err != nil {
		return err
	}

	loader.Resources = append(loader.Resources, resource)
	return nil
}
