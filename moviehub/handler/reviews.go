package handler

import (
 "net/http"
 "errors"
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
