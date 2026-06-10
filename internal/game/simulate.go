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

type circuit struct {
	name        string
	circuitType string  // "normal" | "street" | "technical" | "wet"
	variance    float64 // race-to-race randomness
	dnfBoost    float64 // extra DNF probability
}

var raceCalendar = []circuit{
	{"Bahrain Grand Prix", "normal", 0.05, 0.00},
	{"Saudi Arabian Grand Prix", "street", 0.10, 0.02},
	{"Australian Grand Prix", "technical", 0.08, 0.01},
	{"Japanese Grand Prix", "technical", 0.06, 0.01},
	{"Chinese Grand Prix", "normal", 0.08, 0.01},
	{"Miami Grand Prix", "street", 0.10, 0.01},
	{"Emilia Romagna Grand Prix", "technical", 0.07, 0.01},
	{"Monaco Grand Prix", "street", 0.20, 0.02},
	{"Canadian Grand Prix", "street", 0.12, 0.02},
	{"Spanish Grand Prix", "normal", 0.06, 0.00},
	{"Austrian Grand Prix", "normal", 0.07, 0.01},
	{"British Grand Prix", "normal", 0.08, 0.01},
	{"Hungarian Grand Prix", "technical", 0.07, 0.00},
	{"Belgian Grand Prix", "wet", 0.14, 0.02},
	{"Dutch Grand Prix", "technical", 0.07, 0.01},
	{"Italian Grand Prix", "normal", 0.06, 0.01},
	{"Azerbaijan Grand Prix", "street", 0.18, 0.03},
	{"Singapore Grand Prix", "street", 0.16, 0.02},
	{"United States Grand Prix", "normal", 0.09, 0.01},
	{"Mexico City Grand Prix", "normal", 0.08, 0.01},
	{"São Paulo Grand Prix", "wet", 0.14, 0.02},
	{"Las Vegas Grand Prix", "street", 0.14, 0.02},
	{"Qatar Grand Prix", "normal", 0.07, 0.01},
	{"Abu Dhabi Grand Prix", "normal", 0.07, 0.00},
}

const baseDNFChance = 0.05

// componentWeights defines how much each team role contributes per circuit type.
// Weights sum to 1.0 within each circuit type.
var componentWeights = map[string]map[string]float64{
	"principal": {"normal": 0.30, "technical": 0.20, "street": 0.20, "wet": 0.30},
	"td":        {"normal": 0.20, "technical": 0.40, "street": 0.15, "wet": 0.15},
	"engineer":  {"normal": 0.20, "technical": 0.20, "street": 0.20, "wet": 0.40},
	"chassis":   {"normal": 0.25, "technical": 0.25, "street": 0.30, "wet": 0.20},
	"strategy":  {"normal": 0.10, "technical": 0.10, "street": 0.20, "wet": 0.40},
}

// SimulateSeason runs a full 24-race season for the given picks.
// Driver contribution is the average pace of both drivers.
// Component contribution varies by circuit type.
func SimulateSeason(picks []f1.Pick, fieldAverage float64) SeasonResult {
	if len(picks) == 0 {
		return emptySeasonResult()
	}
	if math.IsNaN(fieldAverage) || math.IsInf(fieldAverage, 0) || fieldAverage <= 0 {
		fieldAverage = 50
	}

	// Average driver pace.
	driverPace := 0.0
	driverCount := 0
	heroDriver := ""
	bestPace := -1.0
	for _, p := range picks {
		if p.Type == "component" {
			continue
		}
		if math.IsNaN(p.Driver.PaceScore) || math.IsInf(p.Driver.PaceScore, 0) {
			continue
		}
		driverPace += p.Driver.PaceScore
		driverCount++
		if p.Driver.PaceScore > bestPace {
			bestPace = p.Driver.PaceScore
			heroDriver = p.Driver.Name
		}
	}
	if driverCount > 0 {
		driverPace /= float64(driverCount)
	} else {
		driverPace = fieldAverage
	}

	// Index component picks by category.
	byCategory := map[string]f1.Pick{}
	for _, p := range picks {
		if p.Type == "component" && p.Component != nil {
			byCategory[p.Component.Category] = p
		}
	}

	races := make([]f1.RaceResult, 0, len(raceCalendar))
	cumWins := 0

	for _, c := range raceCalendar {
		ct := c.circuitType

		// Component sub-score weighted by circuit type.
		compPace := 0.0
		totalCompW := 0.0
		for cat, wMap := range componentWeights {
			w := wMap[ct]
			if w == 0 {
				continue
			}
			var score float64
			if p, ok := byCategory[cat]; ok && p.Component != nil {
				score = p.Component.Score
			} else {
				score = fieldAverage
			}
			compPace += score * w
			totalCompW += w
		}
		if totalCompW > 0 {
			compPace /= totalCompW
		} else {
			compPace = fieldAverage
		}

		// Blend 60% drivers / 40% components.
		teamPace := driverPace*0.60 + compPace*0.40
		winProb := clamp(teamPace/(teamPace+fieldAverage), 0, 0.94)

		isDNF := rand.Float64() < (baseDNFChance + c.dnfBoost)
		won := false
		if !isDNF {
			raceProb := clamp(winProb+(rand.Float64()-0.5)*c.variance*2, 0, 0.99)
			won = rand.Float64() < raceProb
		}
		if won {
			cumWins++
		}

		r := f1.RaceResult{
			Race:        c.name,
			Won:         won,
			DNF:         isDNF,
			Cumulative:  cumWins,
			CircuitType: c.circuitType,
		}
		if won {
			r.HeroDriver = heroDriver
		}
		races = append(races, r)
	}

	return SeasonResult{Wins: cumWins, Races: races}
}

// Simulate is a backward-compatible wrapper used by tests.
// Returns 0 if no driver has a valid PaceScore.
func Simulate(lineup []f1.Driver, fieldAverage float64) int {
	valid := 0
	for _, d := range lineup {
		if !math.IsNaN(d.PaceScore) && !math.IsInf(d.PaceScore, 0) {
			valid++
		}
	}
	if valid == 0 {
		return 0
	}
	picks := make([]f1.Pick, len(lineup))
	for i, d := range lineup {
		picks[i] = f1.Pick{Driver: d}
	}
	return SimulateSeason(picks, fieldAverage).Wins
}

func emptySeasonResult() SeasonResult {
	races := make([]f1.RaceResult, len(raceCalendar))
	for i, c := range raceCalendar {
		races[i] = f1.RaceResult{Race: c.name, CircuitType: c.circuitType}
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
