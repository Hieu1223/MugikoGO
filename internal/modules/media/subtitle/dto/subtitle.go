package dto

import (
	"time"

	"github.com/google/uuid"
)

type ProviderKind string

const (
	ProviderKindRemote ProviderKind = "remote"
	ProviderKindYouTube ProviderKind = "youtube"
)

// Provider DTOs
type SubtitleProviderCreate struct {
	Name     string       `json:"name"`
	Kind     ProviderKind `json:"kind"`
	BaseURL  string       `json:"base_url,omitempty"`
}

type SubtitleProviderServiceResponse struct {
	ID        uuid.UUID `json:"id"`
	ProviderID uuid.UUID `json:"provider_id"`
	ServiceKey string    `json:"service_key"`
	Name      string    `json:"name"`
	Options   any       `json:"options,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SubtitleProviderResponse struct {
	ID            uuid.UUID                      `json:"id"`
	Name          string                         `json:"name"`
	Kind          ProviderKind                   `json:"kind"`
	BaseURL       string                         `json:"base_url,omitempty"`
	IsActive      bool                           `json:"is_active"`
	LastSyncedAt  *time.Time                     `json:"last_synced_at,omitempty"`
	CreatedAt     time.Time                      `json:"created_at"`
	UpdatedAt     time.Time                      `json:"updated_at"`
	Services      []SubtitleProviderServiceResponse `json:"services"`
}

// Subtitle DTOs
type SubtitleCreate struct {
	VideoURL    string    `json:"video_url"`
	ProviderID  uuid.UUID `json:"provider_id"`
	ServiceKey  string    `json:"service_key"`
}

type SubtitleSegmentResponse struct {
	ID        uuid.UUID `json:"id"`
	SubtitleID uuid.UUID `json:"subtitle_id"`
	Idx       int       `json:"idx"`
	StartMs   int       `json:"start_ms"`
	EndMs     int       `json:"end_ms"`
	Words     any       `json:"words,omitempty"`
}

type SubtitleResponse struct {
	ID         uuid.UUID `json:"id"`
	VideoID    string    `json:"video_id"`
	VideoURL   string    `json:"video_url"`
	ProviderID uuid.UUID `json:"provider_id"`
	ServiceID  uuid.UUID `json:"service_id"`
	CreatedBy  uuid.UUID `json:"created_by"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type SubtitleSegmentsResponse struct {
	Items  []SubtitleSegmentResponse `json:"items"`
	Total  int                       `json:"total"`
	Limit  int                       `json:"limit"`
	Offset int                       `json:"offset"`
}

type SubtitleListResponse struct {
	ID            uuid.UUID `json:"id"`
	VideoID       string    `json:"video_id"`
	VideoURL      string    `json:"video_url"`
	ProviderID    uuid.UUID `json:"provider_id"`
	ServiceID     uuid.UUID `json:"service_id"`
	CreatedBy     uuid.UUID `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	SegmentCount  int       `json:"segment_count"`
}

// Settings DTOs
type UserSettingsResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	SchemaVersion int      `json:"schema_version"`
	Data         any       `json:"data"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserSettingsUpdate struct {
	SchemaVersion int  `json:"schema_version"`
	Data          any  `json:"data"`
}

// User State DTOs
type SubtitleUserStateResponse struct {
	UserID        uuid.UUID `json:"user_id"`
	SubtitleID    uuid.UUID `json:"subtitle_id"`
	SchemaVersion int       `json:"schema_version"`
	Data          any       `json:"data"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SubtitleUserStateUpdate struct {
	SchemaVersion int  `json:"schema_version"`
	Data          any  `json:"data"`
}

// Subtitle History (visit history) DTOs
type SubtitleHistoryResponse struct {
	SubtitleID    uuid.UUID `json:"subtitle_id"`
	VideoID       string    `json:"video_id"`
	VideoURL      string    `json:"video_url"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
	SegmentCount  int       `json:"segment_count"`
}

type SubtitleHistoryListResponse struct {
	Items      []SubtitleHistoryResponse `json:"items"`
	Total      int                       `json:"total"`
	Limit      int                       `json:"limit"`
	Offset     int                       `json:"offset"`
}