// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newGenDocsCmd() *cobra.Command {
	var (
		format = "markdown"
		dir    = "."
	)

	formats := map[string]func(cmd *cobra.Command, dir string) error{
		"markdown": doc.GenMarkdownTree,
		"yaml":     doc.GenYamlTree,
		"rest":     doc.GenReSTTree,
	}

	// returns all available format options
	options := func() []string {
		fmts := make([]string, 0, len(formats))
		for fmt := range formats {
			fmts = append(fmts, fmt)
		}
		return fmts
	}

	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate CLI docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			gen, ok := formats[format]
			if !ok {
				return fmt.Errorf("unsupported documentation format. use on of %v", options())
			}

			rootCmd, err := newRootCmd(io.Discard, nil)
			if err != nil {
				return err
			}

			rootCmd.DisableAutoGenTag = true

			if err := gen(rootCmd, dir); err != nil {
				return err
			}

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVar((*string)(&format), "format", format, fmt.Sprintf("Docs format. One of: %v.", options()))
	f.StringVarP((*string)(&dir), "output", "o", dir, "Output directory.")

	return cmd
}
