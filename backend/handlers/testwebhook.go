package handlers

import (
	"net/http"

	vectormgr "vex-backend/vector/manager"
)

// TestHandler returns an http.HandlerFunc that closes over the provided Manager.
// It responds with a simple OK body. The manager is available to the handler
// for future use.
func TestHandler(m vectormgr.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}
