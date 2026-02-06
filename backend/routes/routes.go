package routes

import (
	"net/http"

	"vex-backend/handlers"
	vectormgr "vex-backend/vector/manager"
)

// RegisterRoutes accepts a single Manager instance which is passed into handler constructors.
// This lets us create the embedder/manager once in main and reuse it across handlers.
func RegisterRoutes(m vectormgr.Manager) *http.ServeMux {
	mux := http.NewServeMux()

	// handlers.GitWebhookHandler and handlers.TestHandler are expected to be functions that
	// take a vectormgr.Manager and return an http.HandlerFunc.
	mux.HandleFunc("/git-webhook", handlers.GitWebhookHandler(m))
	mux.HandleFunc("/test", handlers.TestHandler(m))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"vex-backend"}`))
	})

	return mux
}
