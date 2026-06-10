package handlers

import (
	"html/template"

	"github.com/adammcgrogan/24-0/internal/f1"
)

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
}
