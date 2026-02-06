package handlers

import (
	"log"
	"net/http"
	"time"

	"vex-backend/config"
	"vex-backend/git"
	vectormgr "vex-backend/vector/manager"
)

type WebhookPayload struct {
	RepoURL string `json:"repo_url"`
}

// GitWebhookHandler returns an http.HandlerFunc that wraps the original git webhook
// behaviour. It accepts the Manager so you can use it inside the handler.
func GitWebhookHandler(m vectormgr.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Push to notes repo at time %v", time.Now())
		changedFiles, _ := git.GetFiles(config.Config.NotesRepo)
		for _, file := range changedFiles {
			log.Printf("%s", file)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}
}
