package game

import (
	"math"
	"math/rand/v2"

	"github.com/adammcgrogan/24-0/internal/f1"
)

const numRaces = 24
const monteCarloRuns = 1000

// Simulate runs a Monte Carlo season simulation for the given driver lineup
// and returns the expected number of race wins out of 24.
func Simulate(lineup []f1.Driver, fieldAverage float64) int {
	if len(lineup) == 0 {
		return 0
	}

	// Sanitise field average — guard against NaN/Inf from corrupt data.
	if math.IsNaN(fieldAverage) || math.IsInf(fieldAverage, 0) || fieldAverage <= 0 {
		fieldAverage = 50
	}

	// Find best driver by PaceScore; skip NaN values.
	bestIdx := -1
	for i, d := range lineup {
		if math.IsNaN(d.PaceScore) || math.IsInf(d.PaceScore, 0) {
			continue
		}
		if bestIdx == -1 || d.PaceScore > lineup[bestIdx].PaceScore {
			bestIdx = i
		}
	}
	if bestIdx == -1 {
		return 0 // all drivers have invalid pace scores
	}
	best := lineup[bestIdx]

	// Bonus from depth: each additional driver above fieldAverage adds a small lift.
	depthBonus := 0.0
	for i, d := range lineup {
		if i == bestIdx {
			continue
		}
		if !math.IsNaN(d.PaceScore) && d.PaceScore > fieldAverage {
			depthBonus += (d.PaceScore - fieldAverage) * 0.002
		}
	}

	winProb := best.PaceScore / (best.PaceScore + fieldAverage)
	winProb = clamp(winProb+depthBonus, 0, 0.95)

	totalWins := 0
	for range monteCarloRuns {
		wins := 0
		for range numRaces {
			if rand.Float64() < winProb {
				wins++
			}
		}
		totalWins += wins
	}

	return totalWins / monteCarloRuns
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
