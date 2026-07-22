package auth

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	adminEmail        string
	adminPasswordHash string
	jwtSecret         string
}

func NewHandler(adminEmail, adminPasswordHash, jwtSecret string) *Handler {
	return &Handler{
		adminEmail:        adminEmail,
		adminPasswordHash: adminPasswordHash,
		jwtSecret:         jwtSecret,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", h.login)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	validEmail := req.Email == h.adminEmail
	err := bcrypt.CompareHashAndPassword([]byte(h.adminPasswordHash), []byte(req.Password))

	if !validEmail || err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token, err := GenerateToken(h.jwtSecret, req.Email)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
