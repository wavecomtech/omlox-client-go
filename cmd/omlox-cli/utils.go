// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package main

// Returns all IDs from 'ids', except those with names matching 'ignoredIDs'
func filterIDs(ids []string, ignoredIDs []string) []string {
	if ignoredIDs == nil {
		return ids
	}

	var filtered []string
	for _, rel := range ids {
		found := false
		for _, ignoredName := range ignoredIDs {
			if rel == ignoredName {
				found = true
				break
			}
		}
		if !found {
			filtered = append(filtered, rel)
		}
	}

	return filtered
}
