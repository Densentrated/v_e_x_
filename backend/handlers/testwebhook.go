package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	vectormgr "vex-backend/vector/manager"
)

// TestHandler returns an http.HandlerFunc that closes over the provided Manager.
// It accepts a JSON body { "query": "<search text>" } and returns the top 3 results
// as JSON. The handler no longer performs embedding; it only queries the vector DB.
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

		log.Printf("[TestHandler] Running query %q (top 3)", req.Query)
		results, err := m.RetriveNVectorsByQuery(ctx, req.Query, 3)
		if err != nil {
			log.Printf("[TestHandler] query error: %v", err)
			http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("[TestHandler] Query returned %d results", len(results))

		// Prepare response (content, metadata, id)
		type responseItem struct {
			Content  string            `json:"content"`
			Metadata map[string]string `json:"metadata"`
			Id       string            `json:"id"`
		}
		out := make([]responseItem, 0, len(results))
		for _, v := range results {
			out = append(out, responseItem{Content: v.Content, Metadata: v.Metadata, Id: v.Id})
		}

		respBytes, err := json.Marshal(out)
		if err != nil {
			log.Printf("[TestHandler] failed to marshal response: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		log.Printf("[TestHandler] Returning %d results in HTTP response", len(out))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	}
}
