package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"mugiko/internal/modules/core/user/dto"
	"mugiko/internal/modules/core/user/exception"
	"mugiko/internal/modules/core/user/persistent"
	"mugiko/internal/modules/core/user/service"
)

type authContextKey string

const AuthContextKey authContextKey = "auth"

func AuthContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		authService := service.NewAuthService()

		var auth *dto.AuthData
		var newAnonymousToken string

		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			token = strings.TrimSpace(token)

			switch {
			case strings.HasPrefix(token, "user-"):
				firebaseToken := strings.TrimPrefix(token, "user-")
				repo := persistent.NewUserRepository()
				auth, _ = repo.GetUserAuthByAuthID(dto.AuthTypeFirebase, firebaseToken)
				if auth == nil || !auth.IsActive {
					writeError(w, exception.ErrInvalidToken)
					return
				}

			case strings.HasPrefix(token, "admin-"):
				adminToken := strings.TrimPrefix(token, "admin-")
				fmt.Printf("AuthContext: calling VerifyAdminToken with adminToken=%s\n", adminToken)
				auth, _ = authService.VerifyAdminToken(adminToken)
				if auth == nil {
					writeError(w, exception.ErrInvalidToken)
					return
				}

			case strings.HasPrefix(token, "anonymous-"):
				anonToken := strings.TrimPrefix(token, "anonymous-")
				auth, _ = authService.VerifyAnonymousToken(anonToken)
				if auth == nil {
					writeError(w, exception.ErrInvalidToken)
					return
				}

			default:
				fmt.Printf("AuthContext: DEFAULT CASE - token doesn't match any prefix\n")
				writeError(w, exception.ErrInvalidTokenFormat)
				return
			}
		} else {
			// No Bearer token: create new anonymous user
			auth, newAnonymousToken, _ = authService.CreateAnonymousUser()
			if auth == nil {
				writeError(w, exception.ErrInternalServer)
				return
			}
		}

		ctx := context.WithValue(r.Context(), AuthContextKey, auth)
		rw := &responseWriter{ResponseWriter: w}
		rw.Header().Add("X-Auth-Token", newAnonymousToken)
		next.ServeHTTP(rw, r.WithContext(ctx))

		// If we created a new anonymous user, add the token to response header
		if newAnonymousToken != "" {
			rw.Header().Set("X-Auth-Token", newAnonymousToken)
		}
	})
}

type responseWriter struct {
	http.ResponseWriter
}

func GetAuthContext(r *http.Request) *dto.AuthData {
	if auth, ok := r.Context().Value(AuthContextKey).(*dto.AuthData); ok {
		return auth
	}
	return nil
}

type UserInfo struct {
	UserID string
	Role   string
}

func GetUserInfo(ctx context.Context) (UserInfo, bool) {
	if auth, ok := ctx.Value(AuthContextKey).(*dto.AuthData); ok {
		return UserInfo{
			UserID: auth.UserID.String(),
			Role:   string(auth.Role),
		}, true
	}
	return UserInfo{}, false
}

func writeError(w http.ResponseWriter, err *exception.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	w.Write([]byte(`{"error_code":"` + err.Code + `","detail":"` + err.Message + `"}`))
}
