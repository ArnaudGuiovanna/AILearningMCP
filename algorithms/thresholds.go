// Copyright (c) 2026 Arnaud Guiovanna <https://www.aguiovanna.fr>
// GitHub: https://github.com/ArnaudGuiovanna
// SPDX-License-Identifier: MIT

// Package algorithms exposes mastery threshold accessors with a runtime
// bascule between the legacy multi-threshold profile and a unified single
// threshold profile.
//
// Profiles
//
//   - "legacy"  (default):           BKT=0.85, KST=0.70, Mid=0.80
//   - "unified" (REGULATION_THRESHOLD=on):  BKT=KST=Mid=0.85
//
// The bascule is the runtime expression of audit findings F-1.8, F-2.3
// and F-3.10 (multiple incompatible mastery thresholds — see
// docs/audit-report.md and docs/regulation-design/07-threshold-resolver.md).
//
// Promotion of "unified" to default is gated on a fresh
// eval/VERDICT_THRESHOLD_<date>.md matching or exceeding the baseline
// eval/VERDICT.md on V1, V3 and V4.

package algorithms

import "os"

// MasteryBKT returns the threshold for "concept mastered per BKT P(L)".
// Used by the MASTERY_READY alert, mastery_challenge eligibility,
// transfer_challenge gating and the mastered-concept count in
// ComputeMetacognitiveAlerts.
func MasteryBKT() float64 {
	if isUnifiedThreshold() {
		return 0.85
	}
	return 0.85
}

// MasteryKST returns the threshold for "prerequisite considered satisfied
// to unlock a successor in the KST frontier". Used by ComputeFrontier,
// ConceptStatus, OLM cluster classification and cockpit aggregations.
func MasteryKST() float64 {
	if isUnifiedThreshold() {
		return 0.85
	}
	return 0.70
}

// MasteryMid returns the intermediate threshold used by hint-independence
// checks (engine/metacognition.go) and learning_negotiation prereq sanity
// (tools/negotiation.go). Both share the semantic intent "concept solid
// enough that the system expects independent solving" — splitting them is
// YAGNI today (cf docs/regulation-design/07-threshold-resolver.md OQ-7.4).
// In the unified profile this collapses to MasteryBKT.
func MasteryMid() float64 {
	if isUnifiedThreshold() {
		return 0.85
	}
	return 0.80
}

// isUnifiedThreshold reads the REGULATION_THRESHOLD env at every call.
// Strict equality "on" only — "ON", "true", "1" are ignored. Lookup-each-call
// keeps the API testable via t.Setenv with no global reset helper. Cost is
// a single os.Getenv per accessor invocation, negligible at the call rate
// of get_next_activity. Replace with sync.OnceValue if profiling shows hot.
func isUnifiedThreshold() bool {
	return os.Getenv("REGULATION_THRESHOLD") == "on"
}
