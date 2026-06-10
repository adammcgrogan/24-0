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

var tmpl *template.Template

func init() {
	tmpl = template.Must(
		template.New("").Funcs(templateFuncs).ParseFS(templateFS,
			"templates/*.html",
			"templates/partials/*.html",
		),
	)
}

// renderTemplate renders a named template into a buffer first. If rendering
// succeeds the buffer is flushed to the response. On error a generic 500 is
// returned without leaking internal details to the client.
func renderTemplate(w http.ResponseWriter, name string, data any) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("renderTemplate %q: %v", name, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

// renderPartial renders a named partial template into a buffer for HTMX swaps.
func renderPartial(w http.ResponseWriter, name string, data any) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("renderPartial %q: %v", name, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}
