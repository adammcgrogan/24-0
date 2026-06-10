package game

import (
	"fmt"
	"math/rand/v2"

	"github.com/adammcgrogan/24-0/internal/f1"
)

// eraIndex is pre-built at startup: eraID → list of (constructor,season) pairs
// that have at least 2 drivers. Computing this once avoids O(n) work per spin.
type constructorSeason struct {
	constructor string
	season      int
}

var eraIndex map[string][]constructorSeason        // eraID → valid pairs
var eraPairs map[constructorSeason][]f1.Driver      // (constructor,season) → drivers (copy, safe to sort)

func init() {
	buildIndex(f1.All())
}

func buildIndex(drivers []f1.Driver) {
	// Group drivers by (constructor, season).
	raw := map[constructorSeason][]f1.Driver{}
	for _, d := range drivers {
		cs := constructorSeason{d.Constructor, d.Season}
		raw[cs] = append(raw[cs], d)
	}

	eraIndex = map[string][]constructorSeason{}
	eraPairs = map[constructorSeason][]f1.Driver{}

	for cs, ds := range raw {
		if len(ds) < 1 {
			continue
		}
		// Determine which era this season belongs to.
		eraID := eraForSeason(cs.season)
		if eraID == "" {
			continue
		}
		// Store a sorted copy (by Races desc) so we always pick the two most
		// experienced drivers without mutating the original slice.
		sorted := topTwo(ds)
		eraPairs[cs] = sorted
		// Index every constructor-season that has at least one driver. Sparse
		// eras (e.g. the Classic era, where the dataset records only a single
		// notable driver per team-season) would otherwise be empty and make
		// Spin fail. Single-driver slots are handled by the mirror fallback in
		// Spin, so the era always remains playable.
		eraIndex[eraID] = append(eraIndex[eraID], cs)
	}
}

func eraForSeason(season int) string {
	for _, e := range f1.Eras {
		if season >= e.Start && season <= e.End {
			return e.ID
		}
	}
	return ""
}

// topTwo returns a new slice of at most 2 drivers with the highest Races count.
// It never mutates the input.
func topTwo(ds []f1.Driver) []f1.Driver {
	if len(ds) == 0 {
		return nil
	}
	if len(ds) == 1 {
		return []f1.Driver{ds[0]}
	}
	// Single O(n) pass — find max1 and max2 without sorting.
	first, second := 0, -1
	for i := 1; i < len(ds); i++ {
		if ds[i].Races > ds[first].Races {
			second = first
			first = i
		} else if second == -1 || ds[i].Races > ds[second].Races {
			second = i
		}
	}
	result := []f1.Driver{ds[first]}
	if second != -1 {
		result = append(result, ds[second])
	}
	return result
}

// Spin picks a random constructor from one of the remaining eras.
// If lockedEra is non-nil the spin stays in that era (constructor-skip path).
// Returns an error only when the era has no constructor data at all.
func Spin(drivers []f1.Driver, remaining []f1.Era, lockedEra *f1.Era) (f1.SpinResult, error) {
	era, err := chooseEra(remaining, lockedEra)
	if err != nil {
		return f1.SpinResult{}, err
	}

	pairs := eraIndex[era.ID]
	if len(pairs) == 0 {
		return f1.SpinResult{}, fmt.Errorf("no constructor data for era %q", era.ID)
	}

	cs := pairs[rand.IntN(len(pairs))]
	pair := eraPairs[cs] // already a safe copy with ≤2 drivers

	result := f1.SpinResult{
		Era:         era,
		Constructor: cs.constructor,
		Season:      cs.season,
	}
	if len(pair) >= 1 {
		result.DriverA = pair[0]
	}
	if len(pair) >= 2 {
		result.DriverB = pair[1]
	} else {
		// Only one driver recorded for this slot — mirror DriverA so the
		// template always has two valid (non-zero) choices to display.
		// The player effectively sees the same driver twice; in practice this
		// is rare and only happens with very incomplete historical data.
		result.DriverB = pair[0]
	}
	return result, nil
}

func chooseEra(remaining []f1.Era, locked *f1.Era) (f1.Era, error) {
	if locked != nil {
		return *locked, nil
	}
	if len(remaining) == 0 {
		return f1.Era{}, fmt.Errorf("no remaining eras to spin")
	}
	return remaining[rand.IntN(len(remaining))], nil
}

// componentIndex is pre-built at startup: category → eraID → list of components.
var componentIndex map[string]map[string][]f1.TeamComponent

func init() {
	componentIndex = map[string]map[string][]f1.TeamComponent{}
	for _, c := range f1.AllComponents() {
		if componentIndex[c.Category] == nil {
			componentIndex[c.Category] = map[string][]f1.TeamComponent{}
		}
		componentIndex[c.Category][c.Era] = append(componentIndex[c.Category][c.Era], c)
	}
}

// SpinComponent picks a random pair from the given component category.
// Returns an error if the category has no data.
func SpinComponent(category string) (f1.ComponentSpin, error) {
	byEra, ok := componentIndex[category]
	if !ok || len(byEra) == 0 {
		return f1.ComponentSpin{}, fmt.Errorf("no components for category %q", category)
	}

	// Pick a random era that has at least one component in this category.
	eras := make([]string, 0, len(byEra))
	for e := range byEra {
		if len(byEra[e]) > 0 {
			eras = append(eras, e)
		}
	}
	eraID := eras[rand.IntN(len(eras))]
	options := byEra[eraID]

	// Find the full Era struct for display.
	var spinEra f1.Era
	for _, e := range f1.Eras {
		if e.ID == eraID {
			spinEra = e
			break
		}
	}

	spin := f1.ComponentSpin{Category: category, Era: spinEra}

	if len(options) == 1 {
		spin.OptionA = options[0]
		spin.OptionB = options[0]
	} else {
		// Pick 2 distinct options.
		i := rand.IntN(len(options))
		j := rand.IntN(len(options) - 1)
		if j >= i {
			j++
		}
		spin.OptionA = options[i]
		spin.OptionB = options[j]
	}
	return spin, nil
}
