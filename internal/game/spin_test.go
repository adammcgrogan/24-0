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
	defer buildIndex(f1.All())

	spin, err := Spin()
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

func TestSpinRealDataPlayable(t *testing.T) {
	spin, err := Spin()
	if err != nil {
		t.Fatalf("Spin on real data returned error: %v", err)
	}
	if spin.DriverA.Name == "" || spin.DriverB.Name == "" {
		t.Error("Spin produced an empty driver slot")
	}
	if spin.Constructor == "" {
		t.Error("Spin produced empty constructor")
	}
}

func TestSpinEmptyIndexReturnsError(t *testing.T) {
	buildIndex(nil)
	defer buildIndex(f1.All())

	_, err := Spin()
	if err == nil {
		t.Error("expected error for empty index, got nil")
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
