package dto

import (
	"time"

	"github.com/google/uuid"
)

type SourceEnum string

const (
	SourceSyosetsu SourceEnum = "syosetsu"
)

type NovelDTO struct {
	ID        uuid.UUID  `json:"id"`
	Source    SourceEnum `json:"source"`
	SourceID  string     `json:"source_id"`
	URL       string     `json:"url"`
	Title     string     `json:"title"`
	Author    string     `json:"author,omitempty"`
	Synopsis  string     `json:"synopsis,omitempty"`
	AddedBy   uuid.UUID  `json:"added_by"`
	CreatedAt time.Time  `json:"created_at"`
}

type ChapterDTO struct {
	ID        uuid.UUID `json:"id"`
	NovelID   uuid.UUID `json:"novel_id"`
	Idx       int       `json:"idx"`
	Title     string    `json:"title,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ParagraphDTO struct {
	ID        uuid.UUID `json:"id"`
	ChapterID uuid.UUID `json:"chapter_id"`
	Idx       int       `json:"idx"`
	Text      string    `json:"text"`
	Words     []string  `json:"words"`
}

type PreviewRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type PreviewResponse struct {
	URL      string     `json:"url"`
	Title    string     `json:"title"`
	Author   string     `json:"author,omitempty"`
	Synopsis string     `json:"synopsis,omitempty"`
	Source   SourceEnum `json:"source"`
}

type AddNovelRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type AddNovelResponse struct {
	Novel *NovelDTO `json:"novel"`
	JobID string    `json:"job_id,omitempty"`
}

type PaginatedNovelsResponse struct {
	Novels    []*NovelDTO `json:"novels"`
	Total     int         `json:"total"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
}

type NovelWithChaptersResponse struct {
	*NovelDTO
	Chapters []*ChapterDTO `json:"chapters"`
}

type ChapterResponse struct {
	*ChapterDTO
	Paragraphs []*ParagraphDTO `json:"paragraphs"`
}

type SearchRequest struct {
	Query string `json:"query" validate:"required"`
	Limit int    `json:"limit,omitempty"`
}

type SearchResponse struct {
	Items []*NovelDTO `json:"items"`
	Total int         `json:"total"`
}

type SyncRequest struct {
	Force bool `json:"force,omitempty"`
}