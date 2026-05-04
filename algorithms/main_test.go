// Copyright (c) 2026 Arnaud Guiovanna <https://www.aguiovanna.fr>
// GitHub: https://github.com/ArnaudGuiovanna
// SPDX-License-Identifier: MIT

package algorithms

import (
	"os"
	"testing"
)

// TestMain anchors the algorithms package's test suite to the legacy
// threshold profile by default, regardless of the parent environment.
// Tests that explicitly need the unified profile call
// t.Setenv("REGULATION_THRESHOLD", "on"), which overrides this for the
// duration of the test only.
//
// Without this guard, running the suite under REGULATION_THRESHOLD=on
// (e.g., when CI dispatches the unified validation pass) would make
// legacy-fixtured tests like TestComputeFrontier fail by design — see
// docs/regulation-design/07-threshold-resolver.md §6.5.
func TestMain(m *testing.M) {
	_ = os.Unsetenv("REGULATION_THRESHOLD")
	os.Exit(m.Run())
}
