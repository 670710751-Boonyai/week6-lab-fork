package mw

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "moviehub/auth"
)

func TestRequireLogin(t *testing.T) {
    issuer := &auth.Issuer{Secret: []byte("local-test-secret-longer-than-32-characters"), TTL: time.Hour}
    token, err := issuer.Issue(15, "member")
    if err != nil { t.Fatal(err) }
    next := func(w http.ResponseWriter, r *http.Request) {
        who := WhoFrom(r.Context())
        if who.ID != 15 || who.Role != "member" {
            t.Fatalf("wrong identity: %+v", who)
        }
        w.WriteHeader(http.StatusOK)
    }
    for _, tc := range []struct {
        name string
        header string
        status int
        code string
    }{
        {"no header", "", 401, "LOGIN_REQUIRED"},
        {"missing bearer", token, 401, "LOGIN_REQUIRED"},
        {"tampered token", "Bearer " + token + "x", 401, "TOKEN_INVALID"},
        {"valid token", "Bearer " + token, 200, ""},
    } {
        t.Run(tc.name, func(t *testing.T) {
            request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
            if tc.header != "" { request.Header.Set("Authorization", tc.header) }
            response := httptest.NewRecorder()
            RequireLogin(issuer, next)(response, request)
            if response.Code != tc.status {
                t.Fatalf("got status %d, expected %d: %s", response.Code, tc.status, response.Body.String())
            }
            if tc.code != "" && !strings.Contains(response.Body.String(), tc.code) {
                t.Fatalf("missing error code %s: %s", tc.code, response.Body.String())
            }
        })
    }
}

func TestRequireRole(t *testing.T) {
    issuer := &auth.Issuer{Secret: []byte("local-test-secret-longer-than-32-characters"), TTL: time.Hour}
    memberToken, err := issuer.Issue(7, "member")
    if err != nil { t.Fatal(err) }
    adminToken, err := issuer.Issue(7, "admin")
    if err != nil { t.Fatal(err) }
    next := RequireLogin(issuer, RequireRole("admin", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    for _, tc := range []struct { token string; status int }{
        {memberToken, 403},
        {adminToken, 200},
    } {
        request := httptest.NewRequest(http.MethodGet, "/api/admin/members", nil)
        request.Header.Set("Authorization", "Bearer " + tc.token)
        response := httptest.NewRecorder()
        next(response, request)
        if response.Code != tc.status {
            t.Fatalf("role request got %d, want %d: %s", response.Code, tc.status, response.Body.String())
        }
    }
}
