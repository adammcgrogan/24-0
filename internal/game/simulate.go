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
	variance    float64 // race-to-race randomness factor
	dnfBoost    float64 // extra DNF probability on top of baseDNFChance
}

// raceCalendar is the 24-round season used for simulation (2024 F1 calendar).
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

// driverRoleWeights defines how much each era's driver contributes per circuit type.
// Weights sum to 1.0 within each circuit type across all 5 eras.
var driverRoleWeights = map[string]map[string]float64{
	"normal":    {"classic": 0.25, "turbo": 0.12, "schumacher": 0.40, "hybrid": 0.12, "modern": 0.11},
	"technical": {"classic": 0.15, "turbo": 0.45, "schumacher": 0.13, "hybrid": 0.13, "modern": 0.14},
	"street":    {"classic": 0.15, "turbo": 0.13, "schumacher": 0.13, "hybrid": 0.14, "modern": 0.45},
	"wet":       {"classic": 0.20, "turbo": 0.12, "schumacher": 0.13, "hybrid": 0.45, "modern": 0.10},
}

// componentWeights defines the simulation contribution of each non-driver category
// on each circuit type. These sit ALONGSIDE the driver weights — the two sub-scores
// are averaged together so the combined team pace stays in the 0-100 range.
var componentWeights = map[string]map[string]float64{
	// category → circuit type → contribution weight (within component sub-score)
	"principal": {"normal": 0.30, "technical": 0.20, "street": 0.20, "wet": 0.30},
	"td":        {"normal": 0.20, "technical": 0.40, "street": 0.15, "wet": 0.15},
	"engineer":  {"normal": 0.20, "technical": 0.20, "street": 0.20, "wet": 0.40},
	"chassis":   {"normal": 0.25, "technical": 0.25, "street": 0.30, "wet": 0.20},
	"strategy":  {"normal": 0.10, "technical": 0.10, "street": 0.20, "wet": 0.40},
}

// SimulateSeason runs a single F1 season for the given picks.
// Driver role weights vary by circuit type; component contributions add a second
// dimension — a great Technical Director visibly helps on technical circuits, etc.
func SimulateSeason(picks []f1.Pick, fieldAverage float64) SeasonResult {
	if len(picks) == 0 {
		return emptySeasonResult()
	}

	if math.IsNaN(fieldAverage) || math.IsInf(fieldAverage, 0) || fieldAverage <= 0 {
		fieldAverage = 50
	}

	// Separate driver and component picks.
	byEra := map[string]f1.Pick{}
	byCategory := map[string]f1.Pick{}
	for _, p := range picks {
		if p.Type == "component" && p.Component != nil {
			byCategory[p.Component.Category] = p
		} else if p.Type != "component" {
			byEra[p.Era.ID] = p
		}
	}

	races := make([]f1.RaceResult, 0, len(raceCalendar))
	cumWins := 0

	for _, c := range raceCalendar {
		ct := c.circuitType

		// --- Driver sub-score (weighted by era role, 0-100 scale) ---
		dWeights := driverRoleWeights[ct]
		if dWeights == nil {
			dWeights = driverRoleWeights["normal"]
		}
		driverPace := 0.0
		heroDriver, heroRole := "", ""
		bestContrib := -1.0
		for eraID, w := range dWeights {
			var pace float64
			if p, ok := byEra[eraID]; ok && !math.IsNaN(p.Driver.PaceScore) {
				pace = p.Driver.PaceScore
				contrib := pace * w
				if contrib > bestContrib {
					bestContrib = contrib
					heroDriver = p.Driver.Name
					heroRole = p.Era.Role
				}
			} else {
				pace = fieldAverage
			}
			driverPace += pace * w
		}

		// --- Component sub-score (each category weighted by circuit type) ---
		// Weights within componentWeights sum to 1.0 across categories for a given circuit type.
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
			compPace /= totalCompW // normalise to 0-100 scale
		} else {
			compPace = fieldAverage
		}

		// Blend 60% drivers / 40% components.
		teamPace := driverPace*0.60 + compPace*0.40
		winProb := teamPace / (teamPace + fieldAverage)
		winProb = clamp(winProb, 0, 0.94)

		dnfProb := baseDNFChance + c.dnfBoost
		isDNF := rand.Float64() < dnfProb

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
			r.HeroRole = heroRole
		}
		races = append(races, r)
	}

	return SeasonResult{Wins: cumWins, Races: races}
}

// Simulate is a backward-compatible wrapper used by tests.
// It assigns eras round-robin. Returns 0 if no driver has a valid PaceScore.
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
		era := f1.Eras[i%len(f1.Eras)]
		picks[i] = f1.Pick{Driver: d, Era: era}
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
