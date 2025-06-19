// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
)

// Provisioned by ldflags
var (
	version    string
	commitHash string
)

const (
	// It identifies the application itself, the actual instance needs to be identified via environment
	// and other details.
	appName = "omlox"
)

func warning(format string, v ...any) {
	format = fmt.Sprintf("WARNING: %s\n", format)
	fmt.Fprintf(os.Stderr, format, v...)
}

func main() {
	cmd, err := newRootCmd(os.Stdout, os.Args[1:])
	if err != nil {
		warning("%+v", err)
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
