// Copyright (c) 2026 Arnaud Guiovanna <https://www.aguiovanna.fr>
// GitHub: https://github.com/ArnaudGuiovanna
// SPDX-License-Identifier: MIT

package algorithms

import (
	"math"

	"tutor-mcp/models"
)

// BKTInfoGain returns the expected reduction in entropy of P(L|concept)
// after observing the next response under the BKT generative model.
//
// Formula (binary entropy in bits):
//
//	H(p)             = -p log2(p) - (1-p) log2(1-p)
//	P(correct)       = P(L)·(1-P(S)) + (1-P(L))·P(G)
//	P(L | correct)   = P(L)·(1-P(S)) / P(correct)
//	P(L | incorrect) = P(L)·P(S)     / (1 - P(correct))
//	E[H(post)]       = P(correct)·H(P(L|correct)) +
//	                   (1-P(correct))·H(P(L|incorrect))
//	InfoGain         = H(P(L)) - E[H(post)]
//
// Properties:
//
//   - Range: [0, 1] in bits, peaks near P(L) = 0.5 (max prior uncertainty).
//   - Approaches 0 at the saturation edges (P(L) ≈ 0 or 1) — observing
//     the response of an already-known/already-unknown concept tells us
//     nothing new.
//   - Higher P(S) and P(G) reduce the gain: a slip-prone or guess-prone
//     concept gives noisier observations.
//
// Defensive returns:
//
//   - nil ConceptState                   → 0
//   - any of P(L), P(S), P(G) is NaN     → 0
//   - P(L) saturated (≤0 or ≥1)          → 0
//   - P(correct) saturated               → 0 (avoids div-by-zero)
//
// Pure function — no side effects, no time, no logging. Used by
// [4] ConceptSelector in PhaseDiagnostic; testable in isolation.
func BKTInfoGain(cs *models.ConceptState) float64 {
	if cs == nil {
		return 0
	}
	pL := cs.PMastery
	pS := cs.PSlip
	pG := cs.PGuess
	if math.IsNaN(pL) || math.IsNaN(pS) || math.IsNaN(pG) {
		return 0
	}
	if pL <= 0 || pL >= 1 {
		return 0
	}

	pCorrect := pL*(1-pS) + (1-pL)*pG
	if pCorrect <= 0 || pCorrect >= 1 {
		return 0
	}

	pLgivenCorrect := pL * (1 - pS) / pCorrect
	pLgivenIncorrect := pL * pS / (1 - pCorrect)

	hPrior := binaryEntropy(pL)
	hPostExpected := pCorrect*binaryEntropy(pLgivenCorrect) +
		(1-pCorrect)*binaryEntropy(pLgivenIncorrect)

	ig := hPrior - hPostExpected
	if ig < 0 {
		// Numerical artefact at extreme parameter combinations — the
		// information-theoretic guarantee is ig >= 0. Floor to 0.
		return 0
	}
	return ig
}

// BinaryEntropy computes Shannon binary entropy in bits:
//
//	H(p) = -p log2(p) - (1-p) log2(1-p)
//
// Returns 0 at the saturation edges (p ≤ 0 or p ≥ 1) — a benign
// baseline used both internally by BKTInfoGain and externally by
// engine/phase_fsm.go for the mean-entropy DIAGNOSTIC criterion.
// Pure function.
func BinaryEntropy(p float64) float64 {
	if p <= 0 || p >= 1 {
		return 0
	}
	return -p*math.Log2(p) - (1-p)*math.Log2(1-p)
}

// binaryEntropy is the unexported alias kept for backwards-compat
// with internal call sites in this package.
func binaryEntropy(p float64) float64 { return BinaryEntropy(p) }
