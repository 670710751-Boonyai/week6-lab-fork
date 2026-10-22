package mw

import (
    "context"
    "encoding/json"
    "net/http"
    "strings"

    "moviehub/auth"
)

type Who struct {
    ID   int64
    Role string
}

type ctxKey struct{}

// WhoFrom returns the identity verified by RequireLogin.
func WhoFrom(ctx context.Context) Who {
    who, _ := ctx.Value(ctxKey{}).(Who)
    return who
}

func deny(w http.ResponseWriter, status int, code, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(map[string]any{
        "error": map[string]string{"code": code, "message": msg},
    })
}

// RequireLogin checks a Bearer JWT, then saves the verified member ID and role.
func RequireLogin(iss *auth.Issuer, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
        if !found || token == "" {
            deny(w, http.StatusUnauthorized, "LOGIN_REQUIRED", "กรุณาเข้าสู่ระบบก่อน")
            return
        }
        id, role, err := iss.Parse(token)
        if err != nil {
            deny(w, http.StatusUnauthorized, "TOKEN_INVALID", "ตั๋วใช้ไม่ได้หรือหมดอายุ กรุณาเข้าสู่ระบบใหม่")
            return
        }
        ctx := context.WithValue(r.Context(), ctxKey{}, Who{ID: id, Role: role})
        next(w, r.WithContext(ctx))
    }
}

// RequireRole must be nested inside RequireLogin.
// The role comes from the verified JWT, never from a request body.
func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if WhoFrom(r.Context()).Role != role {
            deny(w, http.StatusForbidden, "FORBIDDEN", "บัญชีนี้ไม่มีสิทธิ์ทำรายการนี้")
            return
        }
        next(w, r)
    }
}
