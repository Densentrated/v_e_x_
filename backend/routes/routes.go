package routes

import (
	"net/http"

	"vex-backend/handlers"
)

func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/git-webhook", handlers.GitWebhookHandler)
	mux.HandleFunc("/test", handlers.TestHandler)

	return mux
}
