package auth

import (
	"net/http"
	"strings"
)

// Middleware returns a decorator that rejects requests without a valid
// Bearer token. It's a plain function, not a struct with heavy DI, since
// a single-admin blog has exactly one thing to check.
func Middleware(secret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || tokenString == "" {
				http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}

			if _, err := ParseToken(secret, tokenString); err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			next(w, r)
		}
	}
}
