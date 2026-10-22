package handler

import (
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "strings"
    "unicode/utf8"

    "moviehub/mw"
    "moviehub/store"
)

func (h *Handler) ListWishlist(w http.ResponseWriter, r *http.Request) {
    who := mw.WhoFrom(r.Context())
    items, err := h.Store.ListWishlist(r.Context(), who.ID)
    if err != nil { internal(w, err); return }
    writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
