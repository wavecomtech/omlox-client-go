// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"io"

	"github.com/wavecomtech/omlox-client-go"
	"github.com/wavecomtech/omlox-client-go/internal/cli"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const deleteTrackablesHelp = `
This command deletes trackables from the Omlox Hub.
`

func newDeleteTrackablesCmd(settings cli.EnvSettings, out io.Writer) *cobra.Command {
	var (
		all       bool
		confirmed bool
	)

	cmd := &cobra.Command{
		Use:   "trackables",
		Short: "Deletes trackables from the Hub",
		Long:  deleteTrackablesHelp,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return compListTrackables(toComplete, args, settings)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newOmloxClient(&settings)
			if err != nil {
				return err
			}

			if all {
				if !confirmed {
					cmd.Printf("Are you sure you want to delete all trackables from %s? [Y/n]\n", settings.OmloxHubAPI)
					if !cli.Ask() {
						cmd.Println("canceled...")
						return nil
					}
				}

				return c.Trackables.DeleteAll(context.Background())
			}

			id, err := uuid.Parse(args[0])
			if err != nil {
				return err
			}

			return c.Trackables.Delete(context.Background(), id)
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&all, "all", "a", false, "Deletes all trackables.")
	f.BoolVarP(&confirmed, "yes", "y", false, "Confirm of the operation.")

	return cmd
}

// Provide dynamic auto-completion for trackable names.
func compListTrackables(toComplete string, ignoredTrackabeNames []string, settings cli.EnvSettings) ([]string, cobra.ShellCompDirective) {
	c, err := omlox.New(settings.OmloxHubAPI)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	trackableUUIDs, err := c.Trackables.IDs(context.Background())
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	trackableIDs := make([]string, len(trackableUUIDs))
	for i, id := range trackableUUIDs {
		trackableIDs[i] = id.String()
	}

	return filterIDs(trackableIDs, ignoredTrackabeNames), cobra.ShellCompDirectiveNoFileComp
}
