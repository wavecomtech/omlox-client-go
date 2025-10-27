// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/wavecomtech/omlox-client-go"
	"github.com/wavecomtech/omlox-client-go/internal/cli"
	"github.com/wavecomtech/omlox-client-go/internal/cli/resource"
)

const updateProviderLocationHelp = `
This command updates location providers locations in the Omlox Hub.
`

func newUpdateProvidersLocationsCmd(settings cli.EnvSettings, out io.Writer) *cobra.Command {
	var files []string

	cmd := &cobra.Command{
		Use:     "providers_locations",
		Aliases: []string{"location_providers_locations"},
		Short:   "Update location providers locations in the Hub",
		Long:    updateProviderLocationHelp,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			var in []io.Reader

			if len(files) > 0 {
				for _, name := range files {
					f, err := os.OpenFile(name, os.O_RDONLY, 0444)
					if err != nil {
						return err
					}
					defer f.Close()

					in = append(in, f)
				}
			} else {
				in = append(in, cmd.InOrStdin())
			}

			loader := resource.Loader[omlox.Location]{
				Resources: make([]omlox.Location, 0),
			}
			for _, r := range in {
				if err := loader.LoadJSON(r); err != nil {
					return err
				}
			}

			c, err := newOmloxClient(&settings)
			if err != nil {
				return err
			}

			for _, p := range loader.Resources {
				err := c.Providers.UpdateLocation(context.Background(), p, p.ProviderID)
				if err != nil {
					return err
				}

				fmt.Fprintf(out, "updated: %v location\n", p.ProviderID)
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.StringArrayVarP(&files, "file", "f", []string{}, "The files that contain the location providers locations to update")

	return cmd
}
