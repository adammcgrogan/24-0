package game

import (
	"testing"

	"github.com/adammcgrogan/24-0/internal/f1"
)

func TestSpinReturnsValidResult(t *testing.T) {
	drivers := []f1.Driver{
		{ID: "driver1", Name: "Alice", Constructor: "RedTeam", Season: 2020, Races: 10, PaceScore: 70},
		{ID: "driver2", Name: "Bob", Constructor: "RedTeam", Season: 2020, Races: 9, PaceScore: 60},
		{ID: "driver3", Name: "Carol", Constructor: "BlueTeam", Season: 2021, Races: 8, PaceScore: 50},
		{ID: "driver4", Name: "Dave", Constructor: "BlueTeam", Season: 2021, Races: 7, PaceScore: 40},
	}
	buildIndex(drivers)
	defer buildIndex(f1.All()) // restore

	remaining := []f1.Era{f1.Eras[4]} // Modern
	spin, err := Spin(drivers, remaining, nil)
	if err != nil {
		t.Fatalf("Spin returned error: %v", err)
	}
	if spin.DriverA.Name == "" {
		t.Error("DriverA has empty name")
	}
	if spin.DriverB.Name == "" {
		t.Error("DriverB has empty name")
	}
}

// TestSpinEveryRealEraPlayable guards against eras becoming unspinnable when
// the dataset records only a single driver per constructor-season (as the
// Classic era does). Every era in the real dataset must yield a valid spin.
func TestSpinEveryRealEraPlayable(t *testing.T) {
	for _, era := range f1.Eras {
		spin, err := Spin(f1.All(), []f1.Era{era}, nil)
		if err != nil {
			t.Errorf("era %q is unspinnable: %v", era.ID, err)
			continue
		}
		if spin.DriverA.Name == "" || spin.DriverB.Name == "" {
			t.Errorf("era %q spin produced an empty driver slot", era.ID)
		}
	}
}

func TestSpinEmptyEraReturnsError(t *testing.T) {
	buildIndex(nil) // empty index
	defer buildIndex(f1.All())

	remaining := []f1.Era{f1.Eras[0]}
	_, err := Spin(nil, remaining, nil)
	if err == nil {
		t.Error("expected error for empty era, got nil")
	}
}

func TestSpinNoRemainingErasReturnsError(t *testing.T) {
	_, err := Spin(f1.All(), nil, nil)
	if err == nil {
		t.Error("expected error for no remaining eras, got nil")
	}
}

func TestTopTwo(t *testing.T) {
	ds := []f1.Driver{
		{Name: "A", Races: 5},
		{Name: "B", Races: 10},
		{Name: "C", Races: 3},
	}
	result := topTwo(ds)
	if len(result) != 2 {
		t.Fatalf("topTwo returned %d drivers, want 2", len(result))
	}
	if result[0].Name != "B" {
		t.Errorf("expected first to be B (most races), got %s", result[0].Name)
	}
	if result[1].Name != "A" {
		t.Errorf("expected second to be A, got %s", result[1].Name)
	}
}
