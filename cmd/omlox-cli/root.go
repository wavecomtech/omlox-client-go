// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/wavecomtech/omlox-client-go"
	"github.com/wavecomtech/omlox-client-go/internal/cli"
	"github.com/wavecomtech/omlox-client-go/internal/cli/log"
)

var globalUsage = `The Omlox Hub CLI tool

Common actions for omlox client:

- omlox get trackables
- omlox sub location_updates
- omlox get trackables -o json > backup.trackables.json
- omlox create trackables < backup.trackables.json
- omlox update trackables < backup.trackables.json

Environment variables:

| Name                 | Description                                                         |
|----------------------|---------------------------------------------------------------------|
| OMLOX_HUB_API        | Omlox hub API endpoint.                                             |
`

func newRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   appName,
		Short: "The Omlox Hub CLI tool",
		Long:  globalUsage,
		Args:  cobra.ExactArgs(0),
	}

	flags := cmd.PersistentFlags()

	settings := cli.New()
	settings.AddFlags(flags)

	flags.Parse(args)

	if settings.Debug {
		setupLogger()
	}

	cmd.AddCommand(
		newVersionCmd(out),
		newGetCmd(*settings, out),
		newCreateCmd(*settings, out),
		newUpdateCmd(*settings, out),
		newDeleteCmd(*settings, out),
		newSubCmd(*settings, out),
		newGenCmd(),
	)

	return cmd, nil
}

// newOmloxClient sets up a new Omlox client with given settings.
func newOmloxClient(settings *cli.EnvSettings) (*omlox.Client, error) {
	opts := make([]omlox.ClientOption, 0)

	if settings.Debug {
		httpClient := omlox.DefaultConfiguration().HTTPClient

		httpClient.Transport = &log.SlogerRoundTripper{
			Logger: slog.Default(),
			Base:   httpClient.Transport,
		}

		opts = append(opts, omlox.WithHTTPClient(httpClient))
	}

	return omlox.New(settings.OmloxHubAPI, opts...)
}

func setupLogger() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(log)
}
