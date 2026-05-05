// Copyright (c) 2026 Arnaud Guiovanna <https://www.aguiovanna.fr>
// GitHub: https://github.com/ArnaudGuiovanna
// SPDX-License-Identifier: MIT

package algorithms

import (
	"math"
	"testing"

	"tutor-mcp/models"
)

// bktState builds a ConceptState with the four BKT-relevant fields set
// for info-gain testing. Only PMastery / PSlip / PGuess matter to
// BKTInfoGain; the rest is left at NewConceptState defaults.
func bktState(pL, pS, pG float64) *models.ConceptState {
	cs := models.NewConceptState("L1", "C")
	cs.PMastery = pL
	cs.PSlip = pS
	cs.PGuess = pG
	return cs
}

func TestBKTInfoGain_PeaksAtPL0_5(t *testing.T) {
	// At P(L)=0.5, the prior entropy is maximum (1 bit). With moderate
	// PSlip/PGuess, info gain should be substantially positive AND
	// strictly greater than at P(L) = 0.1 or 0.9 with the same noise.
	pS, pG := 0.1, 0.2
	g05 := BKTInfoGain(bktState(0.5, pS, pG))
	g01 := BKTInfoGain(bktState(0.1, pS, pG))
	g09 := BKTInfoGain(bktState(0.9, pS, pG))
	if g05 <= g01 {
		t.Errorf("expected ig(0.5)=%f > ig(0.1)=%f", g05, g01)
	}
	if g05 <= g09 {
		t.Errorf("expected ig(0.5)=%f > ig(0.9)=%f", g05, g09)
	}
	if g05 < 0.3 {
		// Sanity floor: at PL=0.5 with low noise, information gain
		// should be a meaningful fraction of the prior bit.
		t.Errorf("expected ig(0.5) >= 0.3 (substantial), got %f", g05)
	}
}

func TestBKTInfoGain_NearZeroAtSaturationLow(t *testing.T) {
	// P(L) just above 0 → prior entropy near 0 → info gain near 0.
	g := BKTInfoGain(bktState(0.01, 0.1, 0.2))
	if g > 0.15 {
		t.Errorf("expected ig near 0 at saturated-low PL, got %f", g)
	}
}

func TestBKTInfoGain_NearZeroAtSaturationHigh(t *testing.T) {
	// P(L) just below 1 → prior entropy near 0 → info gain near 0.
	g := BKTInfoGain(bktState(0.99, 0.1, 0.2))
	if g > 0.15 {
		t.Errorf("expected ig near 0 at saturated-high PL, got %f", g)
	}
}

func TestBKTInfoGain_NonNegative(t *testing.T) {
	for _, pL := range []float64{0.05, 0.2, 0.4, 0.5, 0.6, 0.8, 0.95} {
		for _, pS := range []float64{0.0, 0.1, 0.3} {
			for _, pG := range []float64{0.0, 0.1, 0.3} {
				g := BKTInfoGain(bktState(pL, pS, pG))
				if g < 0 {
					t.Errorf("ig(%.2f, %.2f, %.2f) = %f < 0", pL, pS, pG, g)
				}
			}
		}
	}
}

func TestBKTInfoGain_RespectsSlipGuess(t *testing.T) {
	// Higher noise (P(S), P(G)) should reduce info-gain at fixed P(L)=0.5.
	gLow := BKTInfoGain(bktState(0.5, 0.05, 0.05))
	gMid := BKTInfoGain(bktState(0.5, 0.20, 0.20))
	gHigh := BKTInfoGain(bktState(0.5, 0.40, 0.40))
	if !(gLow > gMid && gMid > gHigh) {
		t.Errorf("expected monotone decrease with noise: low=%.3f mid=%.3f high=%.3f",
			gLow, gMid, gHigh)
	}
}

func TestBKTInfoGain_NilStateReturnsZero(t *testing.T) {
	if g := BKTInfoGain(nil); g != 0 {
		t.Errorf("expected 0 for nil state, got %f", g)
	}
}

func TestBKTInfoGain_NaNReturnsZero(t *testing.T) {
	cases := []struct {
		name string
		cs   *models.ConceptState
	}{
		{"NaN PMastery", bktState(math.NaN(), 0.1, 0.2)},
		{"NaN PSlip", bktState(0.5, math.NaN(), 0.2)},
		{"NaN PGuess", bktState(0.5, 0.1, math.NaN())},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := BKTInfoGain(c.cs)
			if g != 0 {
				t.Errorf("expected 0 on %s, got %f", c.name, g)
			}
		})
	}
}

func TestBKTInfoGain_ExactSaturationReturnsZero(t *testing.T) {
	if g := BKTInfoGain(bktState(0, 0.1, 0.2)); g != 0 {
		t.Errorf("expected 0 at P(L)=0, got %f", g)
	}
	if g := BKTInfoGain(bktState(1, 0.1, 0.2)); g != 0 {
		t.Errorf("expected 0 at P(L)=1, got %f", g)
	}
}

// TestBKTInfoGain_NoiselessIsExactlyOneAtPL0_5 confirms the analytic
// upper bound: with P(S)=P(G)=0, observing the response is perfectly
// informative — info gain equals the prior entropy of 1 bit.
func TestBKTInfoGain_NoiselessIsExactlyOneAtPL0_5(t *testing.T) {
	g := BKTInfoGain(bktState(0.5, 0, 0))
	if math.Abs(g-1.0) > 1e-9 {
		t.Errorf("expected exactly 1 bit at P(L)=0.5 noiseless, got %f", g)
	}
}
