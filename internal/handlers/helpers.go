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
}

// ComputeStats derives season statistics from a slice of race results.
func ComputeStats(races []f1.RaceResult) SeasonStats {
	var s SeasonStats
	if len(races) == 0 {
		return s
	}
	s.Wins = 0
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
	// emptySlots returns a slice of nils to range over for unfilled slots.
	"emptySlots": func(picks []f1.Pick) []struct{} {
		n := 5 - len(picks)
		if n < 0 {
			n = 0
		}
		return make([]struct{}, n)
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
