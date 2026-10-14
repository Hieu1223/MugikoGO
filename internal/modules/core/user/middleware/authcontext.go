package middleware

import (
	"context"
	"mugiko/internal/modules/core/user/persistent"
	"net/http"
)

func AuthContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		auth, error := persistent.GetAuth(token)

		ctx := context.WithValue(r.Context(), "auth", auth)
		next.ServeHTTP(w, r.WithContext(ctx)) // Call the next handler
	})
}
