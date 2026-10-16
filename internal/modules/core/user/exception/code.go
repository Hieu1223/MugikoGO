package exception

import "net/http"

const (
	StatusForbidden          = http.StatusForbidden
	StatusInvalidToken       = http.StatusUnauthorized
	StatusInvalidTokenFormat = http.StatusBadRequest
	StatusUserNotFound       = http.StatusNotFound
	StatusUnauthorized       = http.StatusUnauthorized
	StatusProfileNotFound    = http.StatusNotFound
)
