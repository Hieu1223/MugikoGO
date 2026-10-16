package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"mugiko/internal/common/api"
	"mugiko/internal/modules/core/user/dto"
	"mugiko/internal/modules/core/user/exception"
	"mugiko/internal/modules/core/user/middleware"
	"mugiko/internal/modules/core/user/service"
)

func ListUsers(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	if auth == nil || auth.Role != dto.RoleAdmin {
		api.WriteError(w, exception.MsgForbidden, exception.StatusForbidden)
		return
	}

	userService := service.NewUserService()
	users, err := userService.ListUsers(100, 0)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidToken, exception.StatusInvalidToken)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetUserDetail(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	if auth == nil || auth.Role != dto.RoleAdmin {
		api.WriteError(w, exception.MsgForbidden, exception.StatusForbidden)
		return
	}

	userIDStr := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidTokenFormat, exception.StatusInvalidTokenFormat)
		return
	}

	userService := service.NewUserService()
	userAuth, profile, err := userService.GetUserDetail(userID)
	if err != nil || userAuth == nil {
		api.WriteError(w, exception.MsgUserNotFound, exception.StatusUserNotFound)
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
		api.WriteError(w, exception.MsgForbidden, exception.StatusForbidden)
		return
	}

	userIDStr := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidTokenFormat, exception.StatusInvalidTokenFormat)
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, exception.MsgInvalidTokenFormat, exception.StatusInvalidTokenFormat)
		return
	}

	role := dto.Role(req.Role)
	userService := service.NewUserService()
	if err := userService.UpdateUserRole(userID, role); err != nil {
		api.WriteError(w, exception.MsgUserNotFound, exception.StatusUserNotFound)
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
	if auth == nil || auth.Role != dto.RoleAdmin {
		api.WriteError(w, exception.MsgForbidden, exception.StatusForbidden)
		return
	}

	userIDStr := r.PathValue("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidTokenFormat, exception.StatusInvalidTokenFormat)
		return
	}

	userService := service.NewUserService()
	if err := userService.DeactivateUser(userID); err != nil {
		api.WriteError(w, exception.MsgUserNotFound, exception.StatusUserNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
