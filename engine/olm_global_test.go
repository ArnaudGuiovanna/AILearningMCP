// Copyright (c) 2026 Arnaud Guiovanna <https://www.aguiovanna.fr>
// SPDX-License-Identifier: MIT

package engine

import (
	"testing"
	"time"
)

func TestGlobalOLMSnapshot_TypesCompile(t *testing.T) {
	g := &GlobalOLMSnapshot{
		Streak:     3,
		TotalSolid: 12,
		Domains: []DomainSummary{
			{DomainID: "d1", DomainName: "math", Solid: 5, KSTProgress: 0.6},
		},
		CalibrationHistory:  []TimePoint{{Day: "2026-05-03", Value: -1.2}},
		AutonomyHistory:     []TimePoint{{Day: "2026-05-03", Value: 0.7}},
		SatisfactionHistory: []TimePoint{{Day: "2026-05-03", Value: 3.0}},
		Goals: []GoalProgress{
			{DomainID: "d1", PersonalGoal: "g", Progress: 0.6},
		},
		RecentEvents: []LearnerEvent{
			{At: time.Now().UTC(), Kind: "mastery_threshold", Concept: "x", Message: "x atteint le seuil"},
		},
	}
	if g.TotalSolid != 12 || len(g.Domains) != 1 || len(g.Goals) != 1 {
		t.Errorf("unexpected shape: %+v", g)
	}
}

func TestBuildGlobalOLMSnapshot_AggregatesAcrossDomains(t *testing.T) {
	store, raw := newOLMTestStore(t)
	seedLearner(t, raw, "L1")

	seedDomain(t, raw, "L1", "math",    []string{"a", "b"}, map[string][]string{"b": {"a"}}, false)
	seedDomain(t, raw, "L1", "anglais", []string{"x"},      nil,                              false)
	seedDomain(t, raw, "L1", "piano",   []string{"p"},      nil,                              false)
	seedConceptState(t, store, "L1", "a", 0.90, "review")
	seedConceptState(t, store, "L1", "x", 0.90, "review")

	g, err := BuildGlobalOLMSnapshot(store, "L1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(g.Domains) != 3 {
		t.Fatalf("Domains=%d, want 3", len(g.Domains))
	}
	if g.TotalSolid < 2 {
		t.Errorf("TotalSolid=%d, want >=2 (a + x)", g.TotalSolid)
	}
	if len(g.Goals) != 3 {
		t.Errorf("Goals=%d, want 3", len(g.Goals))
	}
}

func TestBuildGlobalOLMSnapshot_NoDomain_ReturnsEmpty(t *testing.T) {
	store, raw := newOLMTestStore(t)
	seedLearner(t, raw, "L1")

	g, err := BuildGlobalOLMSnapshot(store, "L1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(g.Domains) != 0 || g.TotalSolid != 0 {
		t.Errorf("expected empty global snapshot, got %+v", g)
	}
}

func TestBuildGlobalOLMSnapshot_PopulatesSparklinesAndEvents(t *testing.T) {
	store, raw := newOLMTestStore(t)
	seedLearner(t, raw, "L1")
	seedDomain(t, raw, "L1", "math", []string{"a"}, nil, false)
	seedConceptState(t, store, "L1", "a", 0.90, "review") // → mastery_threshold event

	now := time.Now().UTC()
	// Seed one interaction today so streak >= 1 (and streak_start event fires).
	if _, err := raw.Exec(
		`INSERT INTO interactions (learner_id, concept, activity_type, success, created_at) VALUES (?,?,?,?,?)`,
		"L1", "a", "RECALL", 1, now,
	); err != nil {
		t.Fatal(err)
	}

	g, err := BuildGlobalOLMSnapshot(store, "L1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if g.Streak < 1 {
		t.Errorf("Streak=%d, want >=1", g.Streak)
	}
	if len(g.RecentEvents) == 0 {
		t.Errorf("RecentEvents empty — expected mastery_threshold from p_mastery=0.90")
	}
	// Day format check on whichever sparkline is non-empty (calibration is empty
	// without a calibration_id row, but RecentEvents.At should be a real time).
	if len(g.RecentEvents) > 0 {
		got := g.RecentEvents[0].At
		if got.IsZero() {
			t.Errorf("RecentEvents[0].At zero, want non-zero")
		}
	}
}
