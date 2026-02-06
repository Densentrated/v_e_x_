package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"vex-backend/chat"
	vectormgr "vex-backend/vector/manager"
)

// TestHandler returns an http.HandlerFunc that closes over the provided Manager.
// It accepts a JSON body { "query": "<search text>" } and uses the ProcessQuery function
// to provide intelligent answers based on the knowledge base.
func TestHandler(m vectormgr.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		log.Printf("[TestHandler] invoked from %s", r.RemoteAddr)

		// Parse JSON body: { "query": "..." }
		var req struct {
			Query string `json:"query"`
		}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				http.Error(w, "missing JSON body", http.StatusBadRequest)
				return
			}
			log.Printf("[TestHandler] invalid JSON: %v", err)
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if req.Query == "" {
			http.Error(w, "field 'query' is required", http.StatusBadRequest)
			return
		}

		log.Printf("[TestHandler] Processing query %q", req.Query)
		answer, err := chat.ProcessQuery(ctx, m, req.Query)
		if err != nil {
			log.Printf("[TestHandler] ProcessQuery error: %v", err)
			http.Error(w, "query processing error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("[TestHandler] Generated answer for query")

		// Prepare response with the answer
		response := struct {
			Query  string `json:"query"`
			Answer string `json:"answer"`
		}{
			Query:  req.Query,
			Answer: answer,
		}

		respBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("[TestHandler] failed to marshal response: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		log.Printf("[TestHandler] Returning answer in HTTP response")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	}
}
