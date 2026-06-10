// Precompute fetches driver stats from the Jolpica F1 API for all seasons
// 1950–2024 and writes normalised driver-season data to internal/f1/drivers.json.
//
// NOTE: internal/f1/drivers.json currently contains hand-curated TEST DATA.
// Run this script before production launch to replace it with real API data.
// The Jolpica API rate-limits aggressively; expect ~5–10 minutes for a full run.
//
// Run once: go run ./cmd/precompute/main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const jolpikaBase = "https://api.jolpi.ca/ergast/f1"
const pageSize = 100

type jolpikaDriver struct {
	DriverID   string `json:"driverId"`
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
}

type jolpikaConstructor struct {
	Name string `json:"name"`
}

type jolpikaResult struct {
	Driver      jolpikaDriver      `json:"Driver"`
	Constructor jolpikaConstructor `json:"Constructor"`
	Position    string             `json:"position"`
	Points      string             `json:"points"`
}

type jolpikaRace struct {
	Results []jolpikaResult `json:"Results"`
}

// driverSeason uses json tags matching f1.Driver.
type driverSeason struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Constructor string  `json:"constructor"`
	Season      int     `json:"season"`
	Wins        int     `json:"wins"`
	Poles       int     `json:"poles"`
	Points      float64 `json:"points"`
	Races       int     `json:"races"`
	PaceScore   float64 `json:"pace_score"`
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

func getJSON(url string, dest any) error {
	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*3) * time.Second)
		}
		resp, err := httpClient.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(10 * time.Second)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			return fmt.Errorf("decode: %w", err)
		}
		return nil
	}
	return fmt.Errorf("all retries exhausted")
}

type resultsResponse struct {
	MRData struct {
		Total     string `json:"total"`
		Limit     string `json:"limit"`
		Offset    string `json:"offset"`
		RaceTable struct {
			Races []jolpikaRace `json:"Races"`
		} `json:"RaceTable"`
	} `json:"MRData"`
}

// fetchAllPages retrieves every result for a given season URL prefix using pagination.
func fetchAllPages(baseURL string) ([]jolpikaRace, error) {
	var allRaces []jolpikaRace
	offset := 0
	for {
		url := fmt.Sprintf("%s?limit=%d&offset=%d", baseURL, pageSize, offset)
		var data resultsResponse
		if err := getJSON(url, &data); err != nil {
			return nil, err
		}
		allRaces = append(allRaces, data.MRData.RaceTable.Races...)

		total, _ := strconv.Atoi(data.MRData.Total)
		offset += pageSize
		if offset >= total {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	return allRaces, nil
}

func fetchSeason(year int) ([]driverSeason, error) {
	races, err := fetchAllPages(fmt.Sprintf("%s/%d/results.json", jolpikaBase, year))
	if err != nil {
		return nil, err
	}
	if len(races) == 0 {
		return nil, fmt.Errorf("no races returned")
	}

	type key struct{ driver, constructor string }
	stats := map[key]*driverSeason{}

	for _, race := range races {
		for _, r := range race.Results {
			k := key{r.Driver.DriverID, r.Constructor.Name}
			if _, ok := stats[k]; !ok {
				stats[k] = &driverSeason{
					ID:          r.Driver.DriverID,
					Name:        r.Driver.GivenName + " " + r.Driver.FamilyName,
					Constructor: r.Constructor.Name,
					Season:      year,
				}
			}
			stats[k].Races++
			if r.Position == "1" {
				stats[k].Wins++
			}
			var pts float64
			fmt.Sscanf(r.Points, "%f", &pts)
			stats[k].Points += pts
		}
	}

	// Fetch qualifying for pole counts (paginated)
	qRaces, err := fetchAllPages(fmt.Sprintf("%s/%d/qualifying.json", jolpikaBase, year))
	if err == nil {
		for _, race := range qRaces {
			// qualifying results are embedded in a different field name;
			// the Jolpica API returns them in Results for qualifying endpoint
			for _, r := range race.Results {
				if r.Position == "1" {
					k := key{r.Driver.DriverID, r.Constructor.Name}
					if s, ok := stats[k]; ok {
						s.Poles++
					}
				}
			}
		}
	}

	var out []driverSeason
	for _, s := range stats {
		out = append(out, *s)
	}
	return out, nil
}

func normalise(all []driverSeason) {
	maxPPR := 0.0
	for _, d := range all {
		if d.Races > 0 {
			if ppr := d.Points / float64(d.Races); ppr > maxPPR {
				maxPPR = ppr
			}
		}
	}
	if maxPPR == 0 {
		maxPPR = 1
	}
	for i := range all {
		d := &all[i]
		if d.Races == 0 {
			continue
		}
		wpr := float64(d.Wins) / float64(d.Races)
		polesPR := float64(d.Poles) / float64(d.Races)
		ptsPR := (d.Points / float64(d.Races)) / maxPPR
		// Weighted sum; already 0–100 without extra scaling
		d.PaceScore = (wpr * 40) + (polesPR * 20) + (ptsPR * 40)
	}
}

func main() {
	var all []driverSeason
	for year := 1950; year <= 2024; year++ {
		log.Printf("fetching %d...", year)
		drivers, err := fetchSeason(year)
		if err != nil {
			log.Printf("  skip %d: %v", year, err)
			continue
		}
		log.Printf("  got %d driver entries", len(drivers))
		all = append(all, drivers...)
		time.Sleep(300 * time.Millisecond)
	}

	normalise(all)

	out, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("internal/f1/drivers.json", out, 0644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %d driver-season records to internal/f1/drivers.json", len(all))
}
