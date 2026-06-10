package f1

// Era represents a distinct F1 era (analogous to a decade in 82-0).
type Era struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Role     string `json:"role"`      // e.g. "Race Ace"
	RoleDesc string `json:"role_desc"` // one-line description shown in the draft UI
}

// Eras is the fixed set of five F1 eras, one per draft round.
// Each era corresponds to a team role with a distinct contribution to the season simulation.
var Eras = []Era{
	{
		ID: "classic", Name: "Classic Era", Start: 1950, End: 1979,
		Role:     "Race Ace",
		RoleDesc: "Your all-round anchor. Versatile on every circuit — the backbone of your team.",
	},
	{
		ID: "turbo", Name: "Turbo Era", Start: 1980, End: 1993,
		Role:     "The Qualifier",
		RoleDesc: "Dominates from the front row. Best on fast, technical circuits where qualifying matters most.",
	},
	{
		ID: "schumacher", Name: "Schumacher Era", Start: 1994, End: 2009,
		Role:     "The Champion",
		RoleDesc: "Your primary race winner. Ruthlessly consistent — drives the bulk of your season points.",
	},
	{
		ID: "hybrid", Name: "Hybrid Era", Start: 2010, End: 2019,
		Role:     "The Strategist",
		RoleDesc: "Reads tyres, adapts to safety cars, and thrives in chaotic wet conditions.",
	},
	{
		ID: "modern", Name: "Modern Era", Start: 2020, End: 2024,
		Role:     "The Pacesetter",
		RoleDesc: "Raw speed through street circuits and power tracks. Sets the tempo everyone else chases.",
	},
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

// TeamComponent is a non-driver team element — principal, technical director, etc.
type TeamComponent struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`    // "principal"|"td"|"engineer"|"chassis"|"strategy"
	Era         string  `json:"era"`         // one of the 5 era IDs
	Description string  `json:"description"` // one-line bio / flavour
	Score       float64 `json:"score"`       // 0–100, normalised within category
	Specialty   string  `json:"specialty"`   // circuit type this excels at: "all"|"street"|"technical"|"wet"|"normal"
}

// ComponentSpin is the pending non-driver choice, analogous to SpinResult for drivers.
type ComponentSpin struct {
	Category string        `json:"category"`
	Era      Era           `json:"era"`
	OptionA  TeamComponent `json:"option_a"`
	OptionB  TeamComponent `json:"option_b"`
}

// ComponentCategoryMeta describes one of the 5 non-driver team slots.
type ComponentCategoryMeta struct {
	ID          string
	Name        string
	Description string
}

// ComponentCategories is the ordered list of the 5 non-driver team roles.
var ComponentCategories = []ComponentCategoryMeta{
	{ID: "principal", Name: "Team Principal",     Description: "Sets team culture and keeps everyone pointed at the same goal."},
	{ID: "td",        Name: "Technical Director",  Description: "Designs the car concept — decisive for qualifying and technical circuits."},
	{ID: "engineer",  Name: "Race Engineer",        Description: "Your driver's voice on the radio. Manages setup, keeps strategy on track."},
	{ID: "chassis",   Name: "Chassis",              Description: "The car's fundamental architecture. Affects every single race."},
	{ID: "strategy",  Name: "Head of Strategy",     Description: "Reads tyres, calls the undercut, thrives in chaos."},
}

// Pick records one of the player's ten draft choices.
// Type == "driver" uses Driver; type == "component" uses Component.
// Old sessions have Type == "" — treat as "driver".
type Pick struct {
	Type      string         `json:"type,omitempty"` // "driver" | "component"
	Driver    Driver         `json:"driver,omitempty"`
	Component *TeamComponent `json:"component,omitempty"`
	Era       Era            `json:"era"`
}

// RaceResult records the outcome of one race in the simulated season.
type RaceResult struct {
	Race        string `json:"race"`
	Won         bool   `json:"won"`
	DNF         bool   `json:"dnf"`
	Cumulative  int    `json:"cumulative"`
	CircuitType string `json:"circuit_type"` // "normal" | "street" | "technical" | "wet"
	HeroDriver  string `json:"hero_driver,omitempty"` // driver name who led the win
	HeroRole    string `json:"hero_role,omitempty"`   // their role label
}

// Session holds the in-progress or completed game state.
type Session struct {
	ID                   string         `json:"id"`
	Picks                []Pick         `json:"picks"`
	PendingSpin          *SpinResult    `json:"pending_spin,omitempty"`
	PendingComponentSpin *ComponentSpin `json:"pending_component_spin,omitempty"`
	ConstructorSkipsLeft int            `json:"constructor_skips_left"`
	EraSkipsLeft         int            `json:"era_skips_left"` // kept for DB compat, unused in new UI
	Wins                 int            `json:"wins"`
	Tier                 string         `json:"tier"`
	Completed            bool           `json:"completed"`
	RaceResults          []RaceResult   `json:"race_results,omitempty"`
}

// DriverPicks returns only the driver picks.
func (s *Session) DriverPicks() []Pick {
	var out []Pick
	for _, p := range s.Picks {
		if p.Type != "component" {
			out = append(out, p)
		}
	}
	return out
}

// ComponentPicks returns only the non-driver picks.
func (s *Session) ComponentPicks() []Pick {
	var out []Pick
	for _, p := range s.Picks {
		if p.Type == "component" {
			out = append(out, p)
		}
	}
	return out
}

// NeedsDriver returns true when fewer than 2 drivers have been picked.
func (s *Session) NeedsDriver() bool {
	return len(s.DriverPicks()) < 2
}

// RemainingComponentCategories returns component category IDs not yet picked.
func (s *Session) RemainingComponentCategories() []ComponentCategoryMeta {
	picked := map[string]bool{}
	for _, p := range s.Picks {
		if p.Type == "component" && p.Component != nil {
			picked[p.Component.Category] = true
		}
	}
	var out []ComponentCategoryMeta
	for _, cat := range ComponentCategories {
		if !picked[cat.ID] {
			out = append(out, cat)
		}
	}
	return out
}

// IsComplete returns true when both drivers and all 5 component roles are picked.
func (s *Session) IsComplete() bool {
	return !s.NeedsDriver() && len(s.RemainingComponentCategories()) == 0
}
