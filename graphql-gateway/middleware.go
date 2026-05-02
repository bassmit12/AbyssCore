package main

import (
	"context"
	"net/http"
	"os"
	"strings"
)

type contextKey string

const ctxKeyToken contextKey = "token"

// authMiddleware extracts the Bearer token from the Authorization header
// and injects it into the request context so resolvers can forward it to Encore services.
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		ctx := context.WithValue(r.Context(), ctxKeyToken, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func tokenFromCtx(ctx context.Context) string {
	t, _ := ctx.Value(ctxKeyToken).(string)
	return t
}

// encoreBaseURL returns the Encore backend base URL from env.
func encoreBaseURL() string {
	u := os.Getenv("ENCORE_BASE_URL")
	if u == "" {
		return "http://localhost:4000"
	}
	return strings.TrimSuffix(u, "/")
}
