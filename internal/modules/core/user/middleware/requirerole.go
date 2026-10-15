package middleware

import (
	"net/http"

	"mugiko/internal/modules/core/user/dto"
	"mugiko/internal/modules/core/user/exception"
)

func RequireRole(allowedRoles ...dto.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := GetAuthContext(r)
			if auth == nil {
				writeError(w, exception.ErrUnauthorized)
				return
			}

			allowed := false
			for _, role := range allowedRoles {
				if auth.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				writeError(w, exception.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireUser(next http.Handler) http.Handler {
	return RequireRole(dto.RoleUser, dto.RoleAdmin)(next)
}

func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(dto.RoleAdmin)(next)
}
