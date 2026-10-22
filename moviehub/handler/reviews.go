package handler

import (
 "net/http"
 "errors"
 "encoding/json"
 "strings"
 "unicode/utf8"
 "moviehub/mw"
 "moviehub/store"
)

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
    who := mw.WhoFrom(r.Context())
    m, err := h.Store.MemberByID(r.Context(), who.ID)
    if errors.Is(err, store.ErrNotFound) {
        writeError(w, 401, "LOGIN_REQUIRED", "กรุณาเข้าสู่ระบบก่อน")
        return
    }
    if err != nil { internal(w, err); return }
    writeJSON(w, 200, m)
}

type reviewReq struct {
    Text string `json:"text"`
}

func (h *Handler) AddReview(w http.ResponseWriter, r *http.Request) {
    movieID, ok := pathID(r, "id")
    if !ok {
        writeError(w, 400, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
        return
    }
    var req reviewReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, 400, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
        return
    }
    text := strings.TrimSpace(req.Text)
    if utf8.RuneCountInString(text) < 10 {
        writeError(w, 400, "REVIEW_TOO_SHORT", "รีวิวสั้นเกินไป")
        return
    }
    who := mw.WhoFrom(r.Context())
    review, err := h.Store.AddReview(r.Context(), who.ID, movieID, text)
    switch {
    case errors.Is(err, store.ErrDuplicate):
        writeError(w, 409, "ALREADY_REVIEWED", "คุณรีวิวเรื่องนี้ไปแล้ว")
    case errors.Is(err, store.ErrNoMovie):
        writeError(w, 404, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
    case err != nil:
        internal(w, err)
    default:
        writeJSON(w, 201, map[string]any{"id": review.ID, "text": review.Text, "createdAt": review.CreatedAt})
    }
}
