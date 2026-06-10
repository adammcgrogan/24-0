package handlers

import (
	"bytes"
	"embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed templates
var templateFS embed.FS

// pageTemplates holds one parsed template set per page so that each page's
// {{define "content"}} block is isolated and cannot be overwritten by another
// page's definition (a known pitfall when all pages share one template.Template).
var pageTemplates map[string]*template.Template

// partialTemplates holds partials for standalone HTMX swap responses.
var partialTemplates *template.Template

var pages = []string{
	"home.html",
	"draft.html",
	"simulate.html",
	"result.html",
	"leaderboard.html",
}

func init() {
	pageTemplates = make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		pageTemplates[page] = template.Must(
			template.New("").Funcs(templateFuncs).ParseFS(templateFS,
				"templates/base.html",
				"templates/partials/spin_result.html",
				"templates/partials/component_spin.html",
				"templates/partials/slot.html",
				"templates/partials/lineup.html",
				"templates/"+page,
			),
		)
	}

	partialTemplates = template.Must(
		template.New("").Funcs(templateFuncs).ParseFS(templateFS,
			"templates/partials/spin_result.html",
			"templates/partials/component_spin.html",
			"templates/partials/slot.html",
			"templates/partials/lineup.html",
		),
	)
}

func renderTemplate(w http.ResponseWriter, name string, data any) {
	t, ok := pageTemplates[name]
	if !ok {
		log.Printf("renderTemplate: unknown page %q", name)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		log.Printf("renderTemplate %q: %v", name, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

func renderPartial(w http.ResponseWriter, name string, data any) {
	var buf bytes.Buffer
	if err := partialTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("renderPartial %q: %v", name, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}
