package handlers

import (
	"log"
	"net/http"
	"time"
	"vex-backend/config"
	"vex-backend/git"
)

type WebhookPayload struct {
	RepoURL string `json:"repo_url"`
}

func GitWebhookHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Push to notes repo at time %v", time.Now())
	changedFiles, _ := git.GetFiles(config.Config.NotesRepo)
	for _, file := range changedFiles {
		log.Printf("%s", file)
	}

	// If a VectorManager has been injected into this package, log its approximate document count.
	// Handlers package expects main (or init code) to call SetVectorManager(...) to inject the manager.
	// The concrete type implements GetDocCount() int64 optionally; use a minimal assertion to avoid importing the concrete type here.
	if VectorManager != nil {
		if getter, ok := VectorManager.(interface{ GetDocCount() int64 }); ok {
			log.Printf("VectorManager document count (approx): %d", getter.GetDocCount())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}
