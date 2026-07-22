package post

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
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

// Register wires this handler's routes onto mux. Auth-protected routes are
// wrapped by the caller (see cmd/api/main.go) so this package stays
// unaware of how auth works. askRateLimit wraps only the /ask route, since that's the one route that costs real money per call.
func (h *Handler) Register(mux *http.ServeMux, authMiddleware func(http.HandlerFunc) http.HandlerFunc, askRateLimit func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /api/posts", h.list)
	mux.HandleFunc("GET /api/posts/{slug}", h.getBySlug)
	mux.HandleFunc("POST /api/posts", authMiddleware(h.create))
	mux.HandleFunc("PUT /api/posts/{id}", authMiddleware(h.update))
	mux.HandleFunc("DELETE /api/posts/{id}", authMiddleware(h.delete))
	mux.HandleFunc("GET /api/admin/posts/{id}", authMiddleware(h.getByID))
	mux.HandleFunc("POST /api/posts/{slug}/ask", askRateLimit(h.ask))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	category := r.URL.Query().Get("category")
	all := r.URL.Query().Get("all") == "true"
	posts, err := h.svc.List(r.Context(), tag, category, !all)
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

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	p, err := h.svc.GetByID(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get post")
		slog.Error("get post by id", "error", err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in CreatePostInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var in UpdatePostInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.svc.Update(r.Context(), id, in)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "post not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete post")
		slog.Error("delete post", "error", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
