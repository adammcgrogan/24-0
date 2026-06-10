package game

import (
	"math"
	"math/rand/v2"

	"github.com/adammcgrogan/24-0/internal/f1"
)

// SeasonResult is returned by SimulateSeason.
type SeasonResult struct {
	Wins  int
	Races []f1.RaceResult
}

// circuit describes a single round on the race calendar.
type circuit struct {
	name     string
	variance float64 // added randomness: street circuits are less predictable
	dnfBoost float64 // extra DNF probability on top of the base 5%
}

// raceCalendar is the 24-round season used for simulation, based on the 2024 F1 calendar.
var raceCalendar = []circuit{
	{"Bahrain Grand Prix", 0.05, 0.00},
	{"Saudi Arabian Grand Prix", 0.10, 0.02},
	{"Australian Grand Prix", 0.08, 0.01},
	{"Japanese Grand Prix", 0.06, 0.01},
	{"Chinese Grand Prix", 0.08, 0.01},
	{"Miami Grand Prix", 0.10, 0.01},       // street-ish
	{"Emilia Romagna Grand Prix", 0.07, 0.01},
	{"Monaco Grand Prix", 0.20, 0.02},      // ultimate street circuit — highest variance
	{"Canadian Grand Prix", 0.12, 0.02},    // wall-lined street hybrid
	{"Spanish Grand Prix", 0.06, 0.00},
	{"Austrian Grand Prix", 0.07, 0.01},
	{"British Grand Prix", 0.08, 0.01},
	{"Hungarian Grand Prix", 0.07, 0.00},
	{"Belgian Grand Prix", 0.09, 0.02},
	{"Dutch Grand Prix", 0.07, 0.01},
	{"Italian Grand Prix", 0.06, 0.01},
	{"Azerbaijan Grand Prix", 0.18, 0.03},  // Baku — chaotic street circuit
	{"Singapore Grand Prix", 0.16, 0.02},   // street circuit, night race
	{"United States Grand Prix", 0.09, 0.01},
	{"Mexico City Grand Prix", 0.08, 0.01},
	{"São Paulo Grand Prix", 0.10, 0.02},   // Interlagos — unpredictable weather
	{"Las Vegas Grand Prix", 0.14, 0.02},   // street circuit
	{"Qatar Grand Prix", 0.07, 0.01},
	{"Abu Dhabi Grand Prix", 0.07, 0.00},
}

const baseDNFChance = 0.05

// SimulateSeason runs a single F1 season for the given driver lineup and returns
// race-by-race results. Each call is random — results vary between games.
func SimulateSeason(lineup []f1.Driver, fieldAverage float64) SeasonResult {
	if len(lineup) == 0 {
		return emptySeasonResult()
	}

	if math.IsNaN(fieldAverage) || math.IsInf(fieldAverage, 0) || fieldAverage <= 0 {
		fieldAverage = 50
	}

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
		return emptySeasonResult()
	}
	best := lineup[bestIdx]

	depthBonus := 0.0
	for i, d := range lineup {
		if i == bestIdx {
			continue
		}
		if !math.IsNaN(d.PaceScore) && d.PaceScore > fieldAverage {
			depthBonus += (d.PaceScore - fieldAverage) * 0.002
		}
	}

	baseWinProb := best.PaceScore / (best.PaceScore + fieldAverage)
	baseWinProb = clamp(baseWinProb+depthBonus, 0, 0.92)

	races := make([]f1.RaceResult, 0, len(raceCalendar))
	cumWins := 0

	for _, c := range raceCalendar {
		dnfProb := baseDNFChance + c.dnfBoost
		isDNF := rand.Float64() < dnfProb

		won := false
		if !isDNF {
			raceProb := clamp(baseWinProb+(rand.Float64()-0.5)*c.variance*2, 0, 0.99)
			won = rand.Float64() < raceProb
		}

		if won {
			cumWins++
		}
		races = append(races, f1.RaceResult{
			Race:       c.name,
			Won:        won,
			DNF:        isDNF,
			Cumulative: cumWins,
		})
	}

	return SeasonResult{Wins: cumWins, Races: races}
}

// Simulate is a convenience wrapper that returns only the win count.
// Use SimulateSeason when you also need the per-race breakdown.
func Simulate(lineup []f1.Driver, fieldAverage float64) int {
	return SimulateSeason(lineup, fieldAverage).Wins
}

func emptySeasonResult() SeasonResult {
	races := make([]f1.RaceResult, len(raceCalendar))
	for i, c := range raceCalendar {
		races[i] = f1.RaceResult{Race: c.name}
	}
	return SeasonResult{Wins: 0, Races: races}
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
