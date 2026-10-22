package mw

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestCORSPreflight(t *testing.T) {
    passedToAPI := 0
    next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        passedToAPI++
        w.WriteHeader(http.StatusTeapot)
    })
    handler := CORS([]string{" http://localhost:3000 ", "https://moviehub.example"}, next)

    cases := []struct {
        name, origin string
        allowed bool
    }{
        {"allowed localhost", "http://localhost:3000", true},
        {"allowed second origin", "https://moviehub.example", true},
        {"other port denied", "http://localhost:3001", false},
        {"no origin", "", false},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            r := httptest.NewRequest(http.MethodOptions, "/api/reviews/1", nil)
            if tc.origin != "" {
                r.Header.Set("Origin", tc.origin)
            }
            r.Header.Set("Access-Control-Request-Method", "DELETE")
            r.Header.Set("Access-Control-Request-Headers", "authorization")
            w := httptest.NewRecorder()
            handler.ServeHTTP(w, r)
            if w.Code != http.StatusNoContent {
                t.Fatalf("status %d, want 204", w.Code)
            }
            got := w.Header().Get("Access-Control-Allow-Origin")
            if tc.allowed {
                if got != tc.origin {
                    t.Fatalf("allow origin %q, want %q", got, tc.origin)
                }
                if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
                    t.Fatal("missing allowed methods")
                }
                if w.Header().Get("Access-Control-Allow-Headers") != "Authorization, Content-Type" {
                    t.Fatal("missing Authorization/Content-Type headers")
                }
                if w.Header().Get("Access-Control-Max-Age") != "600" {
                    t.Fatal("missing 600-second preflight cache duration")
                }
                if w.Header().Get("Vary") != "Origin" {
                    t.Fatal("missing Vary Origin")
                }
            } else if got != "" || w.Header().Get("Access-Control-Allow-Headers") != "" ||
                w.Header().Get("Access-Control-Allow-Methods") != "" {
                t.Fatalf("disallowed origin received CORS headers: %v", w.Header())
            }
        })
    }
    if passedToAPI != 0 {
        t.Fatalf("OPTIONS reached API %d times", passedToAPI)
    }
}

func TestCORSNormalRequest(t *testing.T) {
    handler := CORS([]string{"http://localhost:3000"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    r := httptest.NewRequest(http.MethodGet, "/api/me", nil)
    r.Header.Set("Origin", "http://localhost:3000")
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, r)
    if w.Code != 200 || w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
        t.Fatalf("GET status=%d headers=%v", w.Code, w.Header())
    }
}
