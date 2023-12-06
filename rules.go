// Copyright (c) Omlox Client Go Contributors
// SPDX-License-Identifier: MIT

package omlox

// The rule syntax is a simple Boolean expression consisting of AND connected expressions.
// Each Boolean expression is assigned a positive number as priority.
type LocatingRule struct {
	// The conditions of the LocatingRule.
	// Supported properties are: accuracy, provider_id, type, source, floor, speed, timestamp_diff.
	Expression string

	// The priority of the LocatingRule.
	// The higher the value the higher the priority of the rule.
	Priority int
}
