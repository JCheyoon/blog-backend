package post

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type asker interface {
	Ask(ctx context.Context, systemPrompt, question string) (string, error)
}

type Handler struct {
	svc  *Service
	chat asker // nil-safe: if unset, the /ask route responds 503
}

func NewHandler(svc *Service, chat asker) *Handler {
	return &Handler{svc: svc, chat: chat}
}

// Register wires this handler's routes onto mux. All routes are public and
// return published posts only. askRateLimit wraps only the /ask route, since
// that's the one route that costs real money per call.
func (h *Handler) Register(mux *http.ServeMux, askRateLimit func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/posts", h.list)
	mux.HandleFunc("GET /api/posts/{slug}", h.getBySlug)
	mux.HandleFunc("POST /api/posts/{slug}/ask", askRateLimit(h.ask))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	category := r.URL.Query().Get("category")
	posts, err := h.svc.List(r.Context(), tag, category, true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list posts")
		slog.Error("list posts", "error", err)
		return
	}
	writeJSON(w, http.StatusOK, posts)
}

func (h *Handler) getBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	p, err := h.svc.GetBySlug(r.Context(), slug)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get post")
		slog.Error("get post", "error", err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

type askRequest struct {
	Question string `json:"question"`
}

type askResponse struct {
	Answer string `json:"answer"`
}

// ask answers a question grounded strictly in one post's content. The
// system prompt is the guardrail: it tells Claude not to use outside
// knowledge and not to answer questions unrelated to the post, which is
// what keeps this a "blog assistant" instead of a general chatbot riding
// on your API key.
func (h *Handler) ask(w http.ResponseWriter, r *http.Request) {
	if h.chat == nil {
		writeError(w, http.StatusServiceUnavailable, "chat feature is not configured")
		return
	}

	slug := r.PathValue("slug")
	p, err := h.svc.GetBySlug(r.Context(), slug)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load post")
		slog.Error("ask: load post", "error", err)
		return
	}

	var req askRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Question) == "" {
		writeError(w, http.StatusBadRequest, "question is required")
		return
	}

	systemPrompt := "You are a helpful assistant answering questions about a single blog post. " +
		"Only use the post content below to answer - do not use outside knowledge. " +
		"If the question can't be answered from this post, say so clearly instead of guessing. " +
		"Keep answers concise (2-4 sentences unless the question needs more).\n\n" +
		"Post title: " + p.Title + "\n\nPost content:\n" + p.Content

	answer, err := h.chat.Ask(r.Context(), systemPrompt, req.Question)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to get an answer right now")
		slog.Error("ask: anthropic call", "error", err)
		return
	}

	writeJSON(w, http.StatusOK, askResponse{Answer: answer})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
