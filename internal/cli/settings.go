// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package cli

import (
	"os"

	"github.com/spf13/pflag"
)

const (
	DefaultOmloxHubAPI = "localhost:8081" // Flowcate's Deephub default API endpoint
)

type EnvSettings struct {
	// Omlox Hub API endpoint.
	OmloxHubAPI string

	// Debug indicates whether or not the Omlox Client is running in Debug mode.
	Debug bool
}

// New creates a new environment settings loading the environment variables.
func New() *EnvSettings {
	env := &EnvSettings{
		OmloxHubAPI: envOr("OMLOX_HUB_API", DefaultOmloxHubAPI),
	}

	return env
}

func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.OmloxHubAPI, "addr", s.OmloxHubAPI, "omlox hub API endpoint")
	fs.BoolVar(&s.Debug, "debug", s.Debug, "enable debug logging")
}

func envOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}
