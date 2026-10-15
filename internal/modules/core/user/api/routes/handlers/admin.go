package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"mugiko/internal/modules/core/user/middleware"
	"mugiko/internal/modules/core/user/dto"
	"mugiko/internal/modules/core/user/exception"
	"mugiko/internal/modules/core/user/service"
)

func ListUsers(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	if auth == nil || auth.Role != dto.RoleAdmin {
		writeError(w, exception.ErrForbidden)
		return
	}

	userService := service.NewUserService()
	users, err := userService.ListUsers(100, 0)
	if err != nil {
		writeError(w, exception.ErrInvalidToken)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetUserDetail(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	if auth == nil || auth.Role != dto.RoleAdmin {
		writeError(w, exception.ErrForbidden)
		return
	}

	userIDStr := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, exception.ErrInvalidTokenFormat)
		return
	}

	userService := service.NewUserService()
	userAuth, profile, err := userService.GetUserDetail(userID)
	if err != nil || userAuth == nil {
		writeError(w, exception.ErrUserNotFound)
		return
	}

	response := map[string]interface{}{
		"user_id":   userAuth.UserID,
		"role":      userAuth.Role,
		"auth_id":   userAuth.AuthID,
		"is_active": userAuth.IsActive,
		"joined_at": userAuth.JoinedAt,
		"name":      profile.Name,
		"bio":       profile.Bio,
		"email":     profile.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	if auth == nil || auth.Role != dto.RoleAdmin {
		writeError(w, exception.ErrForbidden)
		return
	}

	userIDStr := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, exception.ErrInvalidTokenFormat)
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, exception.ErrInvalidTokenFormat)
		return
	}

	role := dto.Role(req.Role)
	userService := service.NewUserService()
	if err := userService.UpdateUserRole(userID, role); err != nil {
		if appErr, ok := exception.IsAppError(err); ok {
			writeError(w, appErr)
		} else {
			writeError(w, exception.ErrUserNotFound)
		}
		return
	}

	userAuth, profile, _ := userService.GetUserDetail(userID)
	response := map[string]interface{}{
		"user_id":   userAuth.UserID,
		"role":      userAuth.Role,
		"auth_id":   userAuth.AuthID,
		"is_active": userAuth.IsActive,
		"joined_at": userAuth.JoinedAt,
		"name":      profile.Name,
		"bio":       profile.Bio,
		"email":     profile.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func DeactivateUser(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	fmt.Printf("DeactivateUser: auth context = %+v\n", auth)
	if auth == nil || auth.Role != dto.RoleAdmin {
		fmt.Printf("DeactivateUser: auth is nil or not admin, auth=%+v\n", auth)
		writeError(w, exception.ErrForbidden)
		return
	}

	fmt.Printf("DeactivateUser: auth=%+v\n", auth)
	fmt.Printf("DeactivateUser: path=%s\n", r.URL.Path)
	fmt.Printf("DeactivateUser: user_id path value=%s\n", r.PathValue("user_id"))

	userIDStr := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, exception.ErrInvalidTokenFormat)
		return
	}

	userService := service.NewUserService()
	if err := userService.DeactivateUser(userID); err != nil {
		writeError(w, exception.ErrUserNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}