package handlers

import (
	"html/template"
	"strings"

	"github.com/adammcgrogan/24-0/internal/f1"
)

// SeasonStats holds derived statistics computed from a session's race results.
type SeasonStats struct {
	Wins       int
	Losses     int
	DNFs       int
	BestStreak int
	FirstWin   int
	WinPct     int
	KeyWins    []f1.RaceResult // up to 5 notable wins to highlight
}

// ComputeStats derives season statistics from a slice of race results.
func ComputeStats(races []f1.RaceResult) SeasonStats {
	var s SeasonStats
	if len(races) == 0 {
		return s
	}
	s.FirstWin = -1
	streak := 0
	for i, r := range races {
		if r.DNF {
			s.DNFs++
			streak = 0
		} else if r.Won {
			s.Wins++
			streak++
			if streak > s.BestStreak {
				s.BestStreak = streak
			}
			if s.FirstWin == -1 {
				s.FirstWin = i + 1
			}
			// Collect up to 5 key wins — prefer street/technical/wet over normal.
			if len(s.KeyWins) < 5 && (r.CircuitType != "normal" || len(s.KeyWins) < 3) {
				s.KeyWins = append(s.KeyWins, r)
			}
		} else {
			s.Losses++
			streak = 0
		}
	}
	if s.FirstWin == -1 {
		s.FirstWin = 0
	}
	if len(races) > 0 {
		s.WinPct = (s.Wins * 100) / len(races)
	}
	return s
}

// circuitTypeColor returns a CSS color for a circuit type badge.
var circuitTypeColor = map[string]string{
	"street":    "#f59e0b",
	"technical": "#14b8a6",
	"wet":       "#3b82f6",
	"normal":    "#676767",
}

// templateFuncs registers helper functions available in all templates.
var templateFuncs = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"mod": func(a, b int) int { return a % b },
	"map": func(pairs ...any) map[string]any {
		m := map[string]any{}
		for i := 0; i+1 < len(pairs); i += 2 {
			key, _ := pairs[i].(string)
			m[key] = pairs[i+1]
		}
		return m
	},
	// emptyDriverSlots returns empty structs for unfilled driver slots.
	"emptyDriverSlots": func(picks []f1.Pick) []struct{} {
		drivers := 0
		for _, p := range picks {
			if p.Type != "component" {
				drivers++
			}
		}
		n := 2 - drivers
		if n < 0 {
			n = 0
		}
		return make([]struct{}, n)
	},
	// emptyComponentSlots returns empty structs for unfilled component slots.
	"emptyComponentSlots": func(picks []f1.Pick) []struct{} {
		comps := 0
		for _, p := range picks {
			if p.Type == "component" {
				comps++
			}
		}
		n := 5 - comps
		if n < 0 {
			n = 0
		}
		return make([]struct{}, n)
	},
	// categoryName maps a component category ID to its display name.
	"categoryName": func(catID string) string {
		for _, c := range f1.ComponentCategories {
			if c.ID == catID {
				return c.Name
			}
		}
		return catID
	},
	// specialtyLabel returns a readable label + icon for a component specialty.
	"specialtyLabel": func(s string) string {
		switch s {
		case "street":
			return "🏙️ Street"
		case "technical":
			return "🔧 Technical"
		case "wet":
			return "🌧️ Wet"
		case "all":
			return "🏁 All circuits"
		}
		return s
	},
	// shortRaceName strips " Grand Prix" to keep the grid compact.
	"shortRaceName": func(name string) string {
		s := strings.TrimSuffix(name, " Grand Prix")
		if len(s) > 12 {
			s = s[:11] + "…"
		}
		return s
	},
}
