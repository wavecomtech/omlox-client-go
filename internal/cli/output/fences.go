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

type FenceFormater struct {
	Fences []omlox.Fence
}

var _ Writer = (*FenceFormater)(nil)

func (ff *FenceFormater) WriteTable(out io.Writer) error {
	w := tabwriter.NewWriter(out, 10, 1, 3, ' ', 0)

	format := "%v\t%s\t%v\t%v\n"
	if _, err := fmt.Fprintf(w, format, "ID", "NAME", "X", "Y"); err != nil {
		return err
	}

	for _, f := range ff.Fences {
		c := f.Region.Center()
		if _, err := fmt.Fprintf(w, format, f.ID, f.Name, c.X, c.Y); err != nil {
			return err
		}
	}

	return w.Flush()
}

func (ff *FenceFormater) WriteJSON(out io.Writer) error {
	return json.NewEncoder(out).Encode(ff.Fences)
}
