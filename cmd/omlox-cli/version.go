// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newVersionCmd(out io.Writer) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprint(out, getVersion())
			return err
		},
	}

	return versionCmd
}

// getVersion returns a string version information.
func getVersion() string {
	return fmt.Sprintf("%s version %s (%s)\n", appName, version, commitHash)
}
