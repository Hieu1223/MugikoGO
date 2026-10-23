package middleware

import (
	"net/http"

	"mugiko/internal/common/api"
	commonexception "mugiko/internal/common/exception"
	"mugiko/internal/modules/core/user/dto"
	"mugiko/internal/modules/core/user/exception"
)

func RequireRole(allowedRoles ...dto.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := GetAuthContext(r)
			if auth == nil {
				api.WriteError(w, commonexception.MsgUnauthorized, exception.StatusUnauthorized)
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
				api.WriteError(w, commonexception.MsgForbidden, exception.StatusForbidden)
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
