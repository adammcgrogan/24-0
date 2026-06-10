package game

import "github.com/adammcgrogan/24-0/internal/f1"

type Tier struct {
	Name        string
	Description string
	Class       string // CSS class for colour coding
}

var tiers = []struct {
	minWins int
	tier    Tier
}{
	{24, Tier{"24-0 — PERFECT SEASON", "An unbeatable team for the ages.", "tier-perfect"}},
	{23, Tier{"Near Perfect", "One slip from immortality.", "tier-near-perfect"}},
	{19, Tier{"Dominant Season", "A historically great team.", "tier-dominant"}},
	{14, Tier{"Championship Contender", "A genuine title-winning outfit.", "tier-contender"}},
	{9, Tier{"Race Winner", "Capable of multiple victories.", "tier-winner"}},
	{4, Tier{"Points Finisher", "Reliable but not race-winning pace.", "tier-points"}},
	{0, Tier{"Backmarker", "Fighting to survive, not to win.", "tier-backmarker"}},
}

func TierForWins(wins int) Tier {
	for _, t := range tiers {
		if wins >= t.minWins {
			return t.tier
		}
	}
	return tiers[len(tiers)-1].tier
}

// CachedFieldAverage is the mean PaceScore across all drivers, computed once
// at startup from the embedded dataset. Use this in handlers instead of calling
// FieldAverage(f1.All()) on every request.
var CachedFieldAverage float64

func init() {
	CachedFieldAverage = FieldAverage(f1.All())
}

// FieldAverage computes the mean PaceScore across all drivers in the dataset.
func FieldAverage(drivers []f1.Driver) float64 {
	if len(drivers) == 0 {
		return 50
	}
	sum := 0.0
	for _, d := range drivers {
		sum += d.PaceScore
	}
	return sum / float64(len(drivers))
}
