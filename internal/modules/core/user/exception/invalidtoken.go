package exception

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code    string `json:"error_code"`
	Message string `json:"detail"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrInvalidToken       = &AppError{Code: "INVALID_TOKEN", Message: "Invalid or expired token", Status: http.StatusUnauthorized}
	ErrInvalidTokenFormat = &AppError{Code: "INVALID_TOKEN_FORMAT", Message: "Invalid token format", Status: http.StatusUnauthorized}
	ErrFirebaseToken      = &AppError{Code: "FIREBASE_TOKEN_ERROR", Message: "Invalid Firebase token", Status: http.StatusUnauthorized}
	ErrUserAlreadyExists  = &AppError{Code: "USER_ALREADY_EXISTS", Message: "User already exists", Status: http.StatusConflict}
	ErrProfileNotFound    = &AppError{Code: "PROFILE_NOT_FOUND", Message: "Profile not found", Status: http.StatusNotFound}
	ErrUserNotFound       = &AppError{Code: "USER_NOT_FOUND", Message: "User not found", Status: http.StatusNotFound}
	ErrInvalidRole        = &AppError{Code: "INVALID_ROLE", Message: "Invalid role", Status: http.StatusBadRequest}
	ErrUnauthorized       = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized", Status: http.StatusUnauthorized}
	ErrForbidden          = &AppError{Code: "FORBIDDEN", Message: "Access forbidden", Status: http.StatusForbidden}
	ErrAnonymousUser      = &AppError{Code: "ANONYMOUS_USER", Message: "Anonymous users cannot perform this action", Status: http.StatusForbidden}
	ErrUserDeactivated    = &AppError{Code: "USER_DEACTIVATED", Message: "User account is deactivated", Status: http.StatusForbidden}
	ErrInternalServer     = &AppError{Code: "INTERNAL_SERVER_ERROR", Message: "Internal server error", Status: http.StatusInternalServerError}
)

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
