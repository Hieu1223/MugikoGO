package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"mugiko/internal/common/api"
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
		var err error

		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			token = strings.TrimSpace(token)

			switch {
			case strings.HasPrefix(token, "user-"):
				firebaseToken := strings.TrimPrefix(token, "user-")
				repo := persistent.NewUserRepository()
				auth, _ = repo.GetUserAuthByAuthID(dto.AuthTypeFirebase, firebaseToken)
				if auth == nil || !auth.IsActive {
					api.WriteError(w, exception.MsgInvalidToken, exception.StatusInvalidToken)
					return
				}

			case strings.HasPrefix(token, "admin-"):
				adminToken := strings.TrimPrefix(token, "admin-")
				fmt.Printf("AuthContext: calling VerifyAdminToken with adminToken=%s\n", adminToken)
				auth, _ = authService.VerifyAdminToken(adminToken)
				if auth == nil {
					api.WriteError(w, exception.MsgInvalidToken, exception.StatusInvalidToken)
					return
				}

			case strings.HasPrefix(token, "anonymous-"):
				anonToken := strings.TrimPrefix(token, "anonymous-")
				auth, err = authService.VerifyAnonymousToken(anonToken)
				if err != nil {
					fmt.Print(err.Error())
					api.WriteError(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if auth == nil {
					api.WriteError(w, exception.MsgInvalidToken, exception.StatusInvalidToken)
					return
				}

			default:
				fmt.Printf("AuthContext: DEFAULT CASE - token doesn't match any prefix\n")
				api.WriteError(w, exception.MsgInvalidTokenFormat, exception.StatusInvalidTokenFormat)
				return
			}
		} else {
			// No Bearer token: create new anonymous user
			auth, newAnonymousToken, err = authService.CreateAnonymousUser()
			if err != nil {
				api.WriteError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if auth == nil {
				api.WriteError(w, exception.MsgAnonymousAuthNotCreated, http.StatusInternalServerError)
				return
			}
		}

		// Attach auth to context
		ctx := context.WithValue(r.Context(), AuthContextKey, auth)
		r = r.WithContext(ctx)

		// If we created a new anonymous user, add the token to response header
		if newAnonymousToken != "" {
			w.Header().Set("Authorization", newAnonymousToken)
		}

		// Proceed to the next handler
		next.ServeHTTP(w, r)
	})
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
