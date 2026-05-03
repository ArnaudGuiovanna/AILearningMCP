// Copyright (c) 2026 Arnaud Guiovanna <https://www.aguiovanna.fr>
// SPDX-License-Identifier: MIT

package tools

import (
	"strings"
	"testing"
)

func TestPickConcept_NoAuth(t *testing.T) {
	_, deps := setupToolsTest(t)
	res := callTool(t, deps, registerPickConcept, "", "pick_concept", map[string]any{
		"concept": "loops",
	})
	if !res.IsError {
		t.Fatalf("expected auth error")
	}
	if !strings.Contains(resultText(res), "authentication") {
		t.Fatalf("got %q", resultText(res))
	}
}

func TestPickConcept_HappyPath(t *testing.T) {
	store, deps := setupToolsTest(t)
	d := makeOwnerDomain(t, store, "L_owner", "py")

	res := callTool(t, deps, registerPickConcept, "L_owner", "pick_concept", map[string]any{
		"concept":   "b",
		"domain_id": d.ID,
	})
	if res.IsError {
		t.Fatalf("expected success, got %q", resultText(res))
	}

	got, _ := store.GetDomainByID(d.ID)
	if got.PinnedConcept != "b" {
		t.Fatalf("expected pin=b, got %q", got.PinnedConcept)
	}
}

func TestPickConcept_UnknownConceptRejected(t *testing.T) {
	store, deps := setupToolsTest(t)
	d := makeOwnerDomain(t, store, "L_owner", "py")

	res := callTool(t, deps, registerPickConcept, "L_owner", "pick_concept", map[string]any{
		"concept":   "ghost",
		"domain_id": d.ID,
	})
	if !res.IsError {
		t.Fatalf("expected error for unknown concept")
	}
	if !strings.Contains(resultText(res), "ghost") {
		t.Fatalf("error should name the unknown concept, got %q", resultText(res))
	}

	got, _ := store.GetDomainByID(d.ID)
	if got.PinnedConcept != "" {
		t.Fatalf("DB should be unchanged, got pin=%q", got.PinnedConcept)
	}
}

func TestPickConcept_EmptyConceptClearsPin(t *testing.T) {
	store, deps := setupToolsTest(t)
	d := makeOwnerDomain(t, store, "L_owner", "py")
	if err := store.SetPinnedConcept("L_owner", d.ID, "a"); err != nil {
		t.Fatal(err)
	}

	res := callTool(t, deps, registerPickConcept, "L_owner", "pick_concept", map[string]any{
		"concept":   "",
		"domain_id": d.ID,
	})
	if res.IsError {
		t.Fatalf("expected success on clear, got %q", resultText(res))
	}

	got, _ := store.GetDomainByID(d.ID)
	if got.PinnedConcept != "" {
		t.Fatalf("expected clear, got pin=%q", got.PinnedConcept)
	}
}

func TestPickConcept_ForeignDomainRejected(t *testing.T) {
	store, deps := setupToolsTest(t)
	d := makeOwnerDomain(t, store, "L_owner", "py")

	res := callTool(t, deps, registerPickConcept, "L_attacker", "pick_concept", map[string]any{
		"concept":   "a",
		"domain_id": d.ID,
	})
	if !res.IsError {
		t.Fatalf("expected error for foreign learner")
	}

	got, _ := store.GetDomainByID(d.ID)
	if got.PinnedConcept != "" {
		t.Fatalf("foreign call must not modify DB")
	}
}
