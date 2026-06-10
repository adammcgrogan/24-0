package f1

// Era represents a distinct F1 era (analogous to a decade in 82-0).
type Era struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Eras is the fixed set of five F1 eras, one per draft round.
var Eras = []Era{
	{ID: "classic",    Name: "Classic Era",       Start: 1950, End: 1979},
	{ID: "turbo",      Name: "Turbo Era",          Start: 1980, End: 1993},
	{ID: "schumacher", Name: "Schumacher Era",     Start: 1994, End: 2009},
	{ID: "hybrid",     Name: "Hybrid Era",         Start: 2010, End: 2019},
	{ID: "modern",     Name: "Modern Era",         Start: 2020, End: 2024},
}

// Driver represents a single driver's stats for one season.
type Driver struct {
	ID          string  `json:"id"`           // e.g. "hamilton"
	Name        string  `json:"name"`
	Constructor string  `json:"constructor"`
	Season      int     `json:"season"`
	Wins        int     `json:"wins"`
	Poles       int     `json:"poles"`
	Points      float64 `json:"points"`
	Races       int     `json:"races"`
	PaceScore   float64 `json:"pace_score"` // normalised 0–100 across all eras
}

// SpinResult is what the server returns after a spin — one constructor from
// a specific era, with the two drivers who raced for them that season.
type SpinResult struct {
	Era         Era    `json:"era"`
	Constructor string `json:"constructor"`
	Season      int    `json:"season"`
	DriverA     Driver `json:"driver_a"`
	DriverB     Driver `json:"driver_b"`
}

// Pick records the driver a player chose in one round.
type Pick struct {
	Driver Driver `json:"driver"`
	Era    Era    `json:"era"`
}

// Session holds the in-progress or completed game state.
type Session struct {
	ID                   string  `json:"id"`
	Picks                []Pick  `json:"picks"`                  // filled as player drafts
	PendingSpin          *SpinResult `json:"pending_spin,omitempty"` // awaiting a pick
	ConstructorSkipsLeft int     `json:"constructor_skips_left"` // starts at 1
	EraSkipsLeft         int     `json:"era_skips_left"`         // starts at 1
	Wins                 int     `json:"wins"`
	Tier                 string  `json:"tier"`
	Completed            bool    `json:"completed"`
}

// RemainingEras returns the eras that haven't been picked yet.
func (s *Session) RemainingEras() []Era {
	picked := map[string]bool{}
	for _, p := range s.Picks {
		picked[p.Era.ID] = true
	}
	var out []Era
	for _, e := range Eras {
		if !picked[e.ID] {
			out = append(out, e)
		}
	}
	return out
}
