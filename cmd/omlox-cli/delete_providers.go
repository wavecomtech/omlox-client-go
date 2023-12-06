// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"io"

	"github.com/wavecomtech/omlox-client-go"
	"github.com/wavecomtech/omlox-client-go/internal/cli"

	"github.com/spf13/cobra"
)

const deleteProvidersHelp = `
This command deletes location providers from the Omlox Hub.
`

func newDeleteProvidersCmd(settings cli.EnvSettings, out io.Writer) *cobra.Command {
	var (
		all       bool
		confirmed bool
	)

	cmd := &cobra.Command{
		Use:     "providers",
		Aliases: []string{"location_providers"},
		Short:   "Deletes location providers from the Hub",
		Long:    deleteProvidersHelp,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return compListProviders(toComplete, args, settings)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newOmloxClient(&settings)
			if err != nil {
				return err
			}

			if all {
				if !confirmed {
					cmd.Printf("Are you sure you want to delete all providers from %s? [Y/n]\n", settings.OmloxHubAPI)
					if !cli.Ask() {
						cmd.Println("canceled...")
						return nil
					}
				}

				return c.Providers.DeleteAll(context.Background())
			}

			for _, arg := range args {
				err := c.Providers.Delete(context.Background(), arg)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "deleted: %v\n", arg)
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&all, "all", "a", false, "Deletes all location providers.")
	f.BoolVarP(&confirmed, "yes", "y", false, "Confirm of the operation.")

	return cmd
}

// Provide dynamic auto-completion for trackable names.
func compListProviders(toComplete string, ignoredProviderNames []string, settings cli.EnvSettings) ([]string, cobra.ShellCompDirective) {
	c, err := omlox.New(settings.OmloxHubAPI)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	providerIDs, err := c.Providers.IDs(context.Background())
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	return filterIDs(providerIDs, ignoredProviderNames), cobra.ShellCompDirectiveNoFileComp
}
