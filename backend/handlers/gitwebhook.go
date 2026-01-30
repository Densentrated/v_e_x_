package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type GitHubPushEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	} `json:"repository"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
}

func GitWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var event GitHubPushEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Push to %s on branch %s by %s\n",
		event.Repository.FullName, event.Ref, event.Pusher.Name)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Webhook processed for repo: %s", event.Repository.Name)
}
