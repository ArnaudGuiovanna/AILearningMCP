// Copyright (c) 2026 Arnaud Guiovanna <https://www.aguiovanna.fr>
// SPDX-License-Identifier: MIT

// Package engine — Open Learner Model global aggregator.
//
// BuildGlobalOLMSnapshot rolls up across all non-archived domains for a learner
// to power the cockpit's "Modèle global" tab.

package engine

import (
	"fmt"
	"time"

	"tutor-mcp/db"
)

// TimePoint is a daily aggregate sample (for sparklines).
type TimePoint struct {
	Day   string  `json:"day"`   // YYYY-MM-DD UTC
	Value float64 `json:"value"`
}

// DomainSummary is one row of the Savoir column.
type DomainSummary struct {
	DomainID     string  `json:"domain_id"`
	DomainName   string  `json:"domain_name"`
	PersonalGoal string  `json:"personal_goal"`
	Solid        int     `json:"solid"`
	InProgress   int     `json:"in_progress"`
	Fragile      int     `json:"fragile"`
	NotStarted   int     `json:"not_started"`
	KSTProgress  float64 `json:"kst_progress"`
}

// GoalProgress is one row of the Objectifs column.
type GoalProgress struct {
	DomainID     string  `json:"domain_id"`
	PersonalGoal string  `json:"personal_goal"`
	Progress     float64 `json:"progress"` // mirror of KSTProgress
}

// LearnerEvent is one row of the recent events timeline.
type LearnerEvent struct {
	At      time.Time `json:"at"`
	Kind    string    `json:"kind"` // "mastery_threshold"|"calibration_threshold"|"retention_drop"|"streak_start"
	Message string    `json:"message"`
	Concept string    `json:"concept,omitempty"`
}

// GlobalOLMSnapshot is the aggregator payload for the cockpit's global tab.
type GlobalOLMSnapshot struct {
	Streak              int             `json:"streak"`
	TotalSolid          int             `json:"total_solid"`
	Domains             []DomainSummary `json:"domains"`
	CalibrationHistory  []TimePoint     `json:"calibration_history"`
	AutonomyHistory     []TimePoint     `json:"autonomy_history"`
	SatisfactionHistory []TimePoint     `json:"satisfaction_history"`
	Goals               []GoalProgress  `json:"goals"`
	RecentEvents        []LearnerEvent  `json:"recent_events"`
}

// BuildGlobalOLMSnapshot aggregates across all non-archived domains for a
// learner — powers the cockpit's "Modèle global" tab.
func BuildGlobalOLMSnapshot(store *db.Store, learnerID string) (*GlobalOLMSnapshot, error) {
	g := &GlobalOLMSnapshot{}

	domains, err := store.GetDomainsByLearner(learnerID, false /*includeArchived*/)
	if err != nil {
		return nil, fmt.Errorf("global olm: list domains: %w", err)
	}

	for _, d := range domains {
		snap, err := BuildOLMSnapshot(store, learnerID, d.ID)
		if err != nil {
			continue // skip a broken domain rather than fail the whole view
		}
		g.Domains = append(g.Domains, DomainSummary{
			DomainID:     d.ID,
			DomainName:   d.Name,
			PersonalGoal: d.PersonalGoal,
			Solid:        snap.Solid,
			InProgress:   snap.InProgress,
			Fragile:      snap.Fragile,
			NotStarted:   snap.NotStarted,
			KSTProgress:  snap.KSTProgress,
		})
		g.TotalSolid += snap.Solid
		g.Goals = append(g.Goals, GoalProgress{
			DomainID:     d.ID,
			PersonalGoal: d.PersonalGoal,
			Progress:     snap.KSTProgress,
		})
	}

	g.Streak, _ = store.GetActivityStreak(learnerID)

	// Calibration sparkline — last 30 samples.
	if hist, err := store.GetCalibrationBiasHistory(learnerID, 30); err == nil {
		for i, v := range hist {
			day := time.Now().UTC().AddDate(0, 0, -(len(hist)-1-i)).Format("2006-01-02")
			g.CalibrationHistory = append(g.CalibrationHistory, TimePoint{Day: day, Value: v})
		}
	}

	// Autonomy + satisfaction — derived from last 30 affects (newest-first DESC).
	affects, _ := store.GetRecentAffectStates(learnerID, 30)
	for i := len(affects) - 1; i >= 0; i-- {
		af := affects[i]
		day := af.CreatedAt.UTC().Format("2006-01-02")
		g.AutonomyHistory = append(g.AutonomyHistory, TimePoint{Day: day, Value: af.AutonomyScore})
		g.SatisfactionHistory = append(g.SatisfactionHistory, TimePoint{Day: day, Value: float64(af.Satisfaction)})
	}

	// Recent events — past 7 days.
	since := time.Now().UTC().AddDate(0, 0, -7)
	if rawEvents, err := store.GetRecentLearnerEvents(learnerID, since); err == nil {
		for _, re := range rawEvents {
			g.RecentEvents = append(g.RecentEvents, LearnerEvent{
				At: re.At, Kind: re.Kind, Message: re.Message, Concept: re.Concept,
			})
		}
	}

	return g, nil
}
