package persistent

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProviderKind string

const (
	ProviderKindRemote ProviderKind = "remote"
	ProviderKindYouTube ProviderKind = "youtube"
)

type SubtitleProvider struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Kind         ProviderKind `json:"kind"`
	BaseURL      *string    `json:"base_url,omitempty"`
	IsActive     bool       `json:"is_active"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type SubtitleProviderService struct {
	ID        uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	ServiceKey string    `json:"service_key"`
	Name      string    `json:"name"`
	Options   *string   `json:"options,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Subtitle struct {
	ID         uuid.UUID `json:"id"`
	VideoID    string    `json:"video_id"`
	VideoURL   string    `json:"video_url"`
	ProviderID uuid.UUID `json:"provider_id"`
	ServiceID  uuid.UUID `json:"service_id"`
	CreatedBy  uuid.UUID `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type SubtitleSegment struct {
	ID        uuid.UUID `json:"id"`
	SubtitleID uuid.UUID `json:"subtitle_id"`
	Idx       int       `json:"idx"`
	StartMs   int       `json:"start_ms"`
	EndMs     int       `json:"end_ms"`
	Words     *string   `json:"words,omitempty"`
}

type SubtitleAccess struct {
	UserID         uuid.UUID `json:"user_id"`
	SubtitleID     uuid.UUID `json:"subtitle_id"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
}

type UserSettings struct {
	UserID       uuid.UUID `json:"user_id"`
	SchemaVersion int      `json:"schema_version"`
	Data         string    `json:"data"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SubtitleUserState struct {
	UserID        uuid.UUID `json:"user_id"`
	SubtitleID    uuid.UUID `json:"subtitle_id"`
	SchemaVersion int       `json:"schema_version"`
	Data          string    `json:"data"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ErrProviderNotFound error when provider not found
var ErrProviderNotFound = fmt.Errorf("provider not found")

// ErrSubtitleNotFound error when subtitle not found
var ErrSubtitleNotFound = fmt.Errorf("subtitle not found")

// ErrServiceNotFound error when service not found
var ErrServiceNotFound = fmt.Errorf("service not found")