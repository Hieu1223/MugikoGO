package dto

import (
	"github.com/google/uuid"
	"time"
)

// Provider visibility enum
type ProviderVisibility string

const (
	ProviderVisibilityPublic  ProviderVisibility = "public"
	ProviderVisibilityPrivate ProviderVisibility = "private"
)

// MangaProviderDTO represents a manga provider
type MangaProviderDTO struct {
	ID          uuid.UUID          `json:"id"`
	OwnerID     uuid.UUID          `json:"owner_id"`
	Name        string             `json:"name"`
	BaseURL     string             `json:"base_url"`
	Visibility  ProviderVisibility `json:"visibility"`
	IsBuiltin   bool               `json:"is_builtin"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// MangaSubscriptionDTO represents a user's subscription to a provider
type MangaSubscriptionDTO struct {
	UserID      uuid.UUID     `json:"user_id"`
	ProviderID  uuid.UUID     `json:"provider_id"`
	Provider    *MangaProviderDTO `json:"provider,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

// MangaHistoryDTO represents user's reading history
type MangaHistoryDTO struct {
	UserID      uuid.UUID `json:"user_id"`
	ProviderID  uuid.UUID `json:"provider_id"`
	MangaID     string    `json:"manga_id"`
	ChapterID   string    `json:"chapter_id"`
	PageIndex   int       `json:"page_index"`
	Title       string    `json:"title"`
	CoverURL    string    `json:"cover_url,omitempty"`
	LastReadAt  time.Time `json:"last_read_at"`
	Provider    *MangaProviderDTO `json:"provider,omitempty"`
}

// MangaUserSettingsDTO represents user settings
type MangaUserSettingsDTO struct {
	UserID        uuid.UUID `json:"user_id"`
	SchemaVersion int       `json:"schema_version"`
	Data          map[string]interface{} `json:"data"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Built-in provider DTOs
type BuiltinProviderInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	APIVersion  string `json:"api_version"`
}

type BuiltinTagResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Slug     string    `json:"slug"`
	Category string    `json:"category"`
}

type BuiltinTagListResponse struct {
	Items []BuiltinTagResponse `json:"items"`
}

type BuiltinChapterSummary struct {
	ID           uuid.UUID `json:"id"`
	Number       string    `json:"number"`
	PublishedAt  time.Time `json:"published_at"`
}

type BuiltinMangaSummary struct {
	ID                uuid.UUID              `json:"id"`
	Slug              string                 `json:"slug"`
	Title             string                 `json:"title"`
	AlternativeTitles []string               `json:"alternative_titles"`
	CoverURL          string                 `json:"cover_url,omitempty"`
	ThumbnailURL      string                 `json:"thumbnail_url,omitempty"`
	Tags              []BuiltinTagResponse   `json:"tags"`
	Score             float64                `json:"score"`
	Views             int                    `json:"views"`
	LatestChapters    []BuiltinChapterSummary `json:"latest_chapters"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type BuiltinMangaListResponse struct {
	Items      []BuiltinMangaSummary `json:"items"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	Total      int                   `json:"total"`
}

type BuiltinMangaDetail struct {
	ID                uuid.UUID              `json:"id"`
	Slug              string                 `json:"slug"`
	Title             string                 `json:"title"`
	AlternativeTitles []string               `json:"alternative_titles"`
	CoverURL          string                 `json:"cover_url,omitempty"`
	ThumbnailURL      string                 `json:"thumbnail_url,omitempty"`
	Description       string                 `json:"description,omitempty"`
	SourceURL         string                 `json:"source_url,omitempty"`
	Tags              []BuiltinTagResponse   `json:"tags"`
	Score             float64                `json:"score"`
	Views             int                    `json:"views"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	LatestChapters    []BuiltinChapterSummary `json:"latest_chapters"`
}

type BuiltinChapterDetail struct {
	ID          uuid.UUID              `json:"id"`
	MangaID     uuid.UUID              `json:"manga_id"`
	Number      string                 `json:"number"`
	Title       string                 `json:"title,omitempty"`
	PublishedAt time.Time              `json:"published_at"`
	Pages       []BuiltinPageResponse  `json:"pages"`
}

type BuiltinChapterListResponse struct {
	Items []BuiltinChapterSummary `json:"items"`
}

type BuiltinPageResponse struct {
	Index  int      `json:"index"`
	Width  int      `json:"width"`
	Height int      `json:"height"`
	URLs   []string `json:"urls"`
}

type BuiltinPageListResponse struct {
	Width  int                   `json:"width"`
	Height int                   `json:"height"`
	Pages  []BuiltinPageResponse `json:"pages"`
}

// Request DTOs
type CreateProviderRequest struct {
	Name       string             `json:"name" binding:"required"`
	BaseURL    string             `json:"base_url" binding:"required"`
	Visibility ProviderVisibility `json:"visibility"`
}

type UpdateProviderRequest struct {
	Name       *string             `json:"name,omitempty"`
	BaseURL    *string             `json:"base_url,omitempty"`
	Visibility *ProviderVisibility `json:"visibility,omitempty"`
}

type SubscribeRequest struct {
	ProviderID uuid.UUID `json:"provider_id" binding:"required"`
}

type UpsertHistoryRequest struct {
	ProviderID uuid.UUID `json:"provider_id" binding:"required"`
	MangaID    string    `json:"manga_id" binding:"required"`
	ChapterID  string    `json:"chapter_id" binding:"required"`
	PageIndex  int       `json:"page_index"`
	Title      string    `json:"title" binding:"required"`
	CoverURL   string    `json:"cover_url,omitempty"`
}

type UpdateSettingsRequest struct {
	SchemaVersion *int                    `json:"schema_version,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
}