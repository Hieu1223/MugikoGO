package handlers

import (
	"encoding/json"
	"net/http"

	"mugiko/internal/modules/core/user/middleware"
	"mugiko/internal/modules/core/user/exception"
	"mugiko/internal/modules/core/user/service"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	if auth == nil {
		writeError(w, exception.ErrUnauthorized)
		return
	}

	userService := service.NewUserService()
	profile, err := userService.GetMyProfile(auth.UserID)
	if err != nil {
		writeError(w, exception.ErrProfileNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	if auth == nil {
		writeError(w, exception.ErrUnauthorized)
		return
	}

	var req struct {
		Name  *string `json:"name,omitempty"`
		Bio   *string `json:"bio,omitempty"`
		Email *string `json:"email,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, exception.ErrInvalidTokenFormat)
		return
	}

	userService := service.NewUserService()
	profile, err := userService.UpdateMyProfile(auth.UserID, req.Name, req.Bio, req.Email)
	if err != nil {
		writeError(w, exception.ErrProfileNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}