package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"mugiko/internal/modules/core/user/dto"
	"mugiko/internal/modules/core/user/exception"
	"mugiko/internal/modules/core/user/service"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, exception.ErrInvalidTokenFormat)
		return
	}

	authHeader := r.Header.Get("Authorization")
	fmt.Printf("RegisterUser: authHeader=%s\n", authHeader)
	token := extractBearerToken(authHeader)
	fmt.Printf("RegisterUser: token=%s\n", token)
	if token == "" {
		writeError(w, exception.ErrInvalidTokenFormat)
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
		if appErr, ok := exception.IsAppError(err); ok {
			writeError(w, appErr)
		} else {
			writeError(w, exception.ErrInvalidToken)
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+resp.Token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}