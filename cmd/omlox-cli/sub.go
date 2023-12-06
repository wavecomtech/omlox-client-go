// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wavecomtech/omlox-client-go"
	"github.com/wavecomtech/omlox-client-go/internal/cli"
)

const subHelp = `
This command subscribes to read-time events from the Omlox Hub.
Subscriptions are made using named topics.

The Omlox standard supports a few:
	- location_updates
	- collision_events
	- fence_events
	- trackable_motions
	- location_updates:geojson
	- fence_events:geojson

Extra topics can be supported by vendors.
`

func newSubCmd(settings cli.EnvSettings, out io.Writer) *cobra.Command {
	getCmd := &cobra.Command{
		Use:     "subscribe",
		Aliases: []string{"sub"},
		Short:   "Subscribes to real-time events",
		Long:    subHelp,
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return compListTopics(toComplete, args, settings)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

			c, err := newOmloxClient(&settings)
			if err != nil {
				return err
			}

			if err := c.Connect(ctx); err != nil {
				return err
			}

			topic := omlox.Topic(args[0])
			sub, err := c.Subscribe(ctx, topic)
			if err != nil {
				return err
			}

			e := json.NewEncoder(out)

			for updates := range sub.ReceiveRaw() {
				for _, u := range updates.Payload {
					if err := e.Encode(u); err != nil {
						return err
					}
				}
			}

			return c.Close()
		},
	}

	return getCmd
}

// Provide dynamic auto-completion for websockets topics.
func compListTopics(toComplete string, ignoredProviderNames []string, settings cli.EnvSettings) ([]string, cobra.ShellCompDirective) {
	return []string{
		string(omlox.TopicLocationUpdates),
		string(omlox.TopicLocationUpdatesGeoJSON),
		string(omlox.TopicCollisionEvents),
		string(omlox.TopicFenceEvents),
		string(omlox.TopicFenceEventsGeoJSON),
		string(omlox.TopicTrackableMotions),
	}, cobra.ShellCompDirectiveNoFileComp
}
