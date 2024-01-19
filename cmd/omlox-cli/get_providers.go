// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/wavecomtech/omlox-client-go/internal/cli"
	"github.com/wavecomtech/omlox-client-go/internal/cli/output"
)

const getProvidersHelp = `
This command retrieves location providers from the Omlox Hub.
`

func newGetProvidersCmd(settings cli.EnvSettings, out io.Writer) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:     "providers",
		Aliases: []string{"location_providers"},
		Short:   "Retrives location providers from the Hub",
		Long:    getProvidersHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newOmloxClient(&settings)
			if err != nil {
				return err
			}

			providers, err := c.Providers.List(context.Background())
			if err != nil {
				return err
			}

			o, err := output.ParseFormat(format)
			if err != nil {
				return err
			}

			return o.Write(out, &output.ProviderFormater{Providers: providers})
		},
	}

	f := cmd.Flags()
	f.StringVarP((*string)(&format), "output", "o", output.Table.String(), fmt.Sprintf("Output format. One of: %v.", output.Formats()))

	return cmd
}
