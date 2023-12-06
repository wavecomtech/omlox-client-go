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

type ProviderFormater struct {
	Providers []omlox.LocationProvider
}

var _ Writer = (*ProviderFormater)(nil)

func (pf *ProviderFormater) WriteTable(out io.Writer) error {
	w := tabwriter.NewWriter(out, 10, 1, 3, ' ', 0)

	format := "%v\t%s\t%v\t\n"
	if _, err := fmt.Fprintf(w, format, "ID", "NAME", "TYPE"); err != nil {
		return err
	}

	for _, p := range pf.Providers {
		if _, err := fmt.Fprintf(w, format, p.ID, p.Name, p.Type); err != nil {
			return err
		}
	}

	return w.Flush()
}

func (pf *ProviderFormater) WriteJSON(out io.Writer) error {
	return json.NewEncoder(out).Encode(pf.Providers)
}
