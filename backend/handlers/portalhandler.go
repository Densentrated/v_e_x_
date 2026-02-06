package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

var portalTmpl = template.Must(template.ParseFiles(filepath.FromSlash("templates/portal.html")))

// PortalHandler returns an http.HandlerFunc that renders the portal template.
func PortalHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Execute the parsed template (no data to pass, but keep gin.H compatibility if needed)
		if err := portalTmpl.Execute(w, nil); err != nil {
			http.Error(w, "failed to render template", http.StatusInternalServerError)
		}
	}
}
