package f1

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed components.json
var componentsJSON []byte

var allComponents []TeamComponent

func init() {
	if err := json.Unmarshal(componentsJSON, &allComponents); err != nil {
		panic(fmt.Sprintf("failed to load components.json: %v", err))
	}
}

// AllComponents returns the full component dataset.
func AllComponents() []TeamComponent {
	return allComponents
}
