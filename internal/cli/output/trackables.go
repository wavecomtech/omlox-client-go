// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/wavecomtech/omlox-client-go"
)

type TrackableFormater struct {
	Trackables []omlox.Trackable
}

var _ Writer = (*TrackableFormater)(nil)

func (tf *TrackableFormater) WriteTable(out io.Writer) error {
	w := tabwriter.NewWriter(out, 10, 1, 3, ' ', 0)

	format := "%v\t%s\t%v\t%v\n"
	if _, err := fmt.Fprintf(w, format, "ID", "NAME", "TYPE", "LOCATION PROVIDERS"); err != nil {
		return err
	}

	for _, t := range tf.Trackables {
		if _, err := fmt.Fprintf(w, format, t.ID, t.Name, t.Type, providerString(t.LocationProviders)); err != nil {
			return err
		}
	}

	return w.Flush()
}

func (tf *TrackableFormater) WriteJSON(out io.Writer) error {
	return json.NewEncoder(out).Encode(tf.Trackables)
}

func providerString(providers []string) string {
	switch len(providers) {
	case 0:
		return ""
	case 1:
		return providers[0]
	}

	return fmt.Sprintf("%v", providers)
}
