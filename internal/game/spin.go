package game

import (
	"fmt"
	"math/rand/v2"

	"github.com/adammcgrogan/24-0/internal/f1"
)

type constructorSeason struct {
	constructor string
	season      int
}

var allPairs   []constructorSeason
var driverPool map[constructorSeason][]f1.Driver

func init() {
	buildIndex(f1.All())
}

func buildIndex(drivers []f1.Driver) {
	raw := map[constructorSeason][]f1.Driver{}
	for _, d := range drivers {
		cs := constructorSeason{d.Constructor, d.Season}
		raw[cs] = append(raw[cs], d)
	}

	allPairs = nil
	driverPool = map[constructorSeason][]f1.Driver{}
	for cs, ds := range raw {
		if len(ds) < 1 {
			continue
		}
		driverPool[cs] = topTwo(ds)
		allPairs = append(allPairs, cs)
	}
}

// topTwo returns at most 2 drivers with the highest Races count.
func topTwo(ds []f1.Driver) []f1.Driver {
	if len(ds) <= 1 {
		return ds
	}
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

// Spin picks a random constructor-season pair and returns two drivers to choose from.
func Spin() (f1.SpinResult, error) {
	if len(allPairs) == 0 {
		return f1.SpinResult{}, fmt.Errorf("no driver data loaded")
	}
	cs := allPairs[rand.IntN(len(allPairs))]
	pair := driverPool[cs]

	result := f1.SpinResult{
		Constructor: cs.constructor,
		Season:      cs.season,
		DriverA:     pair[0],
	}
	if len(pair) >= 2 {
		result.DriverB = pair[1]
	} else {
		result.DriverB = pair[0]
	}
	return result, nil
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
func SpinComponent(category string) (f1.ComponentSpin, error) {
	byEra, ok := componentIndex[category]
	if !ok || len(byEra) == 0 {
		return f1.ComponentSpin{}, fmt.Errorf("no components for category %q", category)
	}

	eras := make([]string, 0, len(byEra))
	for e := range byEra {
		if len(byEra[e]) > 0 {
			eras = append(eras, e)
		}
	}
	eraID := eras[rand.IntN(len(eras))]
	options := byEra[eraID]

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
