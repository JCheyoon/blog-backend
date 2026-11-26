package category

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/categories", h.tree)
}

// tree is public — returns nested categories for the frontend sidebar.
func (h *Handler) tree(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.Tree(r.Context())
	if err != nil {
		slog.Error("list categories", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list categories")
		return
	}
	if cats == nil {
		cats = []Category{}
	}
	writeJSON(w, http.StatusOK, cats)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
