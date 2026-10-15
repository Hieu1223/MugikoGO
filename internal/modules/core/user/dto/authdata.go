package dto

import (
	"time"

	"github.com/google/uuid"
)

type AuthData struct {
	UserID   uuid.UUID
	Role     Role
	JoinedAt time.Time
	AuthID   string
	IsActive bool
}

type RegisterRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Bio   string `json:"bio,omitempty" validate:"max=500"`
	Email string `json:"email,omitempty"`
}

type RegisterResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	Token  string    `json:"token"`
}

type UserProfileResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	Bio      string    `json:"bio,omitempty"`
	Email    string    `json:"email,omitempty"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type UserProfileUpdate struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Bio   *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	Email *string `json:"email,omitempty"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=anonymous user admin"`
}
