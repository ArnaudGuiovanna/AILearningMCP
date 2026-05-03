// Copyright (c) 2026 Arnaud Guiovanna <https://www.aguiovanna.fr>
// SPDX-License-Identifier: MIT

package engine

import (
	"testing"
)

func TestBuildOLMSnapshot_PinnedConceptOverridesAutoFocus(t *testing.T) {
	store, raw := newOLMTestStore(t)
	seedLearner(t, raw, "L1")
	domainID := seedDomain(t, raw, "L1", "py",
		[]string{"vars", "loops", "funcs"},
		map[string][]string{"loops": {"vars"}, "funcs": {"loops"}},
		false)

	// Pin "funcs" — auto focus would have been "vars" (frontier).
	if err := store.SetPinnedConcept("L1", domainID, "funcs"); err != nil {
		t.Fatal(err)
	}
	snap, err := BuildOLMSnapshot(store, "L1", domainID)
	if err != nil {
		t.Fatal(err)
	}
	if snap.FocusConcept != "funcs" {
		t.Fatalf("expected focus=funcs (pinned), got %q", snap.FocusConcept)
	}
}

func TestBuildOLMSnapshot_PinnedConceptDisappearedFromGraph_SilentClear(t *testing.T) {
	store, raw := newOLMTestStore(t)
	seedLearner(t, raw, "L1")
	domainID := seedDomain(t, raw, "L1", "py",
		[]string{"vars", "loops"},
		nil,
		false)
	if err := store.SetPinnedConcept("L1", domainID, "ghost"); err != nil {
		t.Fatal(err)
	}

	snap, err := BuildOLMSnapshot(store, "L1", domainID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if snap.FocusConcept == "ghost" {
		t.Fatalf("ghost concept should not become focus")
	}
	got, err := store.GetDomainByID(domainID)
	if err != nil {
		t.Fatal(err)
	}
	if got.PinnedConcept != "" {
		t.Fatalf("expected silent clear of stale pin, still %q", got.PinnedConcept)
	}
}

func TestBuildOLMSnapshot_NoPin_FallsBackToAuto(t *testing.T) {
	store, raw := newOLMTestStore(t)
	seedLearner(t, raw, "L1")
	domainID := seedDomain(t, raw, "L1", "py", []string{"vars"}, nil, false)
	snap, err := BuildOLMSnapshot(store, "L1", domainID)
	if err != nil {
		t.Fatal(err)
	}
	if snap.FocusConcept != "vars" {
		t.Fatalf("expected auto focus=vars, got %q", snap.FocusConcept)
	}
}
