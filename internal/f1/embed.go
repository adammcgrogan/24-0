package f1

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed drivers.json
var driversJSON []byte

var allDrivers []Driver

func init() {
	if err := json.Unmarshal(driversJSON, &allDrivers); err != nil {
		panic(fmt.Sprintf("failed to load drivers.json: %v", err))
	}
}

// All returns the full pre-computed driver dataset.
func All() []Driver {
	return allDrivers
}
