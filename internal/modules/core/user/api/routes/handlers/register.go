package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"mugiko/internal/common/api"
	"mugiko/internal/modules/core/user/dto"
	"mugiko/internal/modules/core/user/exception"
	"mugiko/internal/modules/core/user/service"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, exception.MsgInvalidTokenFormat, exception.StatusInvalidTokenFormat)
		return
	}

	authHeader := r.Header.Get("Authorization")
	fmt.Printf("RegisterUser: authHeader=%s\n", authHeader)
	token := api.ExtractBearerToken(authHeader)
	fmt.Printf("RegisterUser: token=%s\n", token)
	if token == "" {
		api.WriteError(w, exception.MsgInvalidTokenFormat, exception.StatusInvalidTokenFormat)
		return
	}

	// For registration, the token is the raw Firebase ID token
	firebaseUser := &service.FirebaseUserInfo{
		UID:   token,
		Email: req.Email,
	}

	authService := service.NewAuthService()
	resp, err := authService.RegisterUser(req, firebaseUser)
	if err != nil {
		// Without a custom error struct, we default to your original fallback.
		// If you want to map specific service errors, you can check err.Error() here:
		// if err.Error() == "user not found" {
		//     api.WriteError(w, exception.MsgUserNotFound, exception.StatusUserNotFound)
		//     return
		// }

		api.WriteError(w, exception.MsgInvalidToken, exception.StatusInvalidToken)
		return
	}

	w.Header().Set("Authorization", "Bearer "+resp.Token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
