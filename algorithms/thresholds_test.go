// Copyright (c) 2026 Arnaud Guiovanna <https://www.aguiovanna.fr>
// GitHub: https://github.com/ArnaudGuiovanna
// SPDX-License-Identifier: MIT

package algorithms

import "testing"

func TestMasteryThresholds_LegacyDefault(t *testing.T) {
	// Explicitly unset so the test is deterministic regardless of the parent
	// environment (CI runners may export REGULATION_THRESHOLD=on for unified
	// validation runs; this test verifies the legacy profile in isolation).
	t.Setenv("REGULATION_THRESHOLD", "")
	if got := MasteryBKT(); got != 0.85 {
		t.Errorf("MasteryBKT legacy: want 0.85, got %v", got)
	}
	if got := MasteryKST(); got != 0.70 {
		t.Errorf("MasteryKST legacy: want 0.70, got %v", got)
	}
	if got := MasteryMid(); got != 0.80 {
		t.Errorf("MasteryMid legacy: want 0.80, got %v", got)
	}
}

func TestMasteryThresholds_Unified(t *testing.T) {
	t.Setenv("REGULATION_THRESHOLD", "on")
	if got := MasteryBKT(); got != 0.85 {
		t.Errorf("MasteryBKT unified: want 0.85, got %v", got)
	}
	if got := MasteryKST(); got != 0.85 {
		t.Errorf("MasteryKST unified: want 0.85, got %v", got)
	}
	if got := MasteryMid(); got != 0.85 {
		t.Errorf("MasteryMid unified: want 0.85, got %v", got)
	}
}

// TestMasteryThresholds_StrictEquality verifies that anything other than
// the exact string "on" leaves the legacy profile active. This protects
// against operator error (e.g. typing "ON" or "true" and silently getting
// the legacy behaviour without warning).
func TestMasteryThresholds_StrictEquality(t *testing.T) {
	bad := []string{"ON", "On", "true", "TRUE", "1", "yes", "enabled", "  on  ", ""}
	for _, v := range bad {
		t.Run("flag="+v, func(t *testing.T) {
			t.Setenv("REGULATION_THRESHOLD", v)
			if MasteryKST() != 0.70 {
				t.Errorf("flag %q must NOT activate unified profile (got KST=%v)", v, MasteryKST())
			}
			if MasteryMid() != 0.80 {
				t.Errorf("flag %q must NOT activate unified profile (got Mid=%v)", v, MasteryMid())
			}
		})
	}
}
