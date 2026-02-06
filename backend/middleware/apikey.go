package middleware

import (
	"net/http"
	"strings"

	"vex-backend/config"
)

// RequireAPIKey is an HTTP middleware that enforces a single hard-coded API key
// defined in config.Config.HardCodedAPIKeyForNow.
//
// The middleware accepts the key via either:
//   - X-API-Key: <key>
//   - Authorization: Bearer <key>
//
// If the configured key is empty or missing, requests will be rejected with
// 401 Unauthorized. If the provided key doesn't match the configured value,
// the request is rejected with 401 Unauthorized.
func RequireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := ""
		if config.Config != nil {
			expected = config.Config.HardCodedAPIKeyForNow
		}

		// If there's no key configured, treat as unauthorized.
		if strings.TrimSpace(expected) == "" {
			http.Error(w, "api key not configured", http.StatusUnauthorized)
			return
		}

		// Try X-API-Key header first.
		key := strings.TrimSpace(r.Header.Get("X-API-Key"))

		// Fallback to Authorization: Bearer <token>
		if key == "" {
			auth := strings.TrimSpace(r.Header.Get("Authorization"))
			if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
				key = strings.TrimSpace(auth[len("Bearer "):])
			}
		}

		// Compare the provided key to the expected key.
		if key == "" || key != expected {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// All good — call the next handler.
		next.ServeHTTP(w, r)
	})
}
