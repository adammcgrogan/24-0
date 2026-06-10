package game

import "testing"

func TestTierForWins(t *testing.T) {
	cases := []struct {
		wins int
		want string
	}{
		{24, "24-0 — PERFECT SEASON"},
		{23, "Near Perfect"},
		{19, "Dominant Season"},
		{14, "Championship Contender"},
		{9, "Race Winner"},
		{4, "Points Finisher"},
		{0, "Backmarker"},
		{1, "Backmarker"},
		{3, "Backmarker"},
	}
	for _, c := range cases {
		got := TierForWins(c.wins).Name
		if got != c.want {
			t.Errorf("TierForWins(%d) = %q, want %q", c.wins, got, c.want)
		}
	}
}
