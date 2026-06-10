package game

import (
	"math"
	"testing"

	"github.com/adammcgrogan/24-0/internal/f1"
)

func TestSimulateBounds(t *testing.T) {
	lineup := []f1.Driver{
		{PaceScore: 80},
		{PaceScore: 70},
		{PaceScore: 60},
		{PaceScore: 50},
		{PaceScore: 40},
	}
	wins := Simulate(lineup, 30)
	if wins < 0 || wins > 24 {
		t.Errorf("Simulate() = %d, want 0–24", wins)
	}
}

func TestSimulateEmpty(t *testing.T) {
	wins := Simulate(nil, 50)
	if wins != 0 {
		t.Errorf("Simulate(nil) = %d, want 0", wins)
	}
}

func TestSimulatePerfect(t *testing.T) {
	lineup := []f1.Driver{{PaceScore: 99}}
	wins := Simulate(lineup, 1)
	if wins < 20 {
		t.Errorf("Simulate(pace=99, field=1) = %d, want >= 20", wins)
	}
}

func TestSimulateNaNPaceScore(t *testing.T) {
	lineup := []f1.Driver{
		{PaceScore: math.NaN()},
		{PaceScore: math.NaN()},
	}
	wins := Simulate(lineup, 50)
	if wins != 0 {
		t.Errorf("Simulate(NaN scores) = %d, want 0", wins)
	}
}

func TestSimulateNaNFieldAverage(t *testing.T) {
	lineup := []f1.Driver{{PaceScore: 70}}
	wins := Simulate(lineup, math.NaN())
	if wins < 0 || wins > 24 {
		t.Errorf("Simulate(NaN fieldAvg) = %d, want 0–24", wins)
	}
}

func TestFieldAverage(t *testing.T) {
	drivers := []f1.Driver{{PaceScore: 10}, {PaceScore: 30}}
	avg := FieldAverage(drivers)
	if avg != 20 {
		t.Errorf("FieldAverage = %f, want 20", avg)
	}
}

func TestFieldAverageEmpty(t *testing.T) {
	avg := FieldAverage(nil)
	if avg != 50 {
		t.Errorf("FieldAverage(nil) = %f, want 50", avg)
	}
}
