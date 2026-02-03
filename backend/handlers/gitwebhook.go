package handlers

import (
	"log"
	"net/http"
	"time"
	"vex-backend/config"
	"vex-backend/storage"
)

func GitWebhookHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Push to notes repo at time %v", time.Now())
	changedFiles, _ := storage.GetFiles(config.Config.NotesRepo)
	for _, file := range changedFiles {
		log.Printf("%s", file)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success"}`))
}
