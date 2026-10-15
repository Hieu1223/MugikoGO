package handlers

import (
	"encoding/json"
	"net/http"

	"mugiko/internal/modules/core/user/exception"
)

func writeError(w http.ResponseWriter, err *exception.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(map[string]string{
		"error_code": err.Code,
		"detail":     err.Message,
	})
}

func extractBearerToken(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}