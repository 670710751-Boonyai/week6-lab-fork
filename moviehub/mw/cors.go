package mw

import (
    "net/http"
    "strings"
)

// CORS allows only the origins listed in ALLOWED_ORIGINS.
// It handles preflight requests before method-specific API routes or JWT checks.
func CORS(allowed []string, next http.Handler) http.Handler {
    ok := make(map[string]bool, len(allowed))
    for _, origin := range allowed {
        origin = strings.TrimSpace(origin)
        if origin != "" {
            ok[origin] = true
        }
    }

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        if origin != "" && ok[origin] {
            h := w.Header()
            h.Set("Access-Control-Allow-Origin", origin)
            h.Add("Vary", "Origin")
            h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
            h.Set("Access-Control-Max-Age", "600")
        }

        // Browsers first send OPTIONS before requests with an Authorization header.
        // Both allowed and unknown origins get 204, but only allowed origins
        // receive the Access-Control-Allow-* response headers.
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}
