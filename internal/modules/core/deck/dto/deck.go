package dto

import (
	"time"

	"github.com/google/uuid"
)

type Deck struct {
	ID        uuid.UUID `json:"id"`
	OwnerID   uuid.UUID `json:"owner_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeckCreateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type DeckUpdateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type DeckResponse struct {
	ID        uuid.UUID `json:"id"`
	OwnerID   uuid.UUID `json:"owner_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeckWord struct {
	ID        uuid.UUID `json:"id"`
	DeckID    uuid.UUID `json:"deck_id"`
	Word      string    `json:"word"`
	Meaning   string    `json:"meaning"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeckWordCreateRequest struct {
	Word    string `json:"word" validate:"required,min=1,max=255"`
	Meaning string `json:"meaning" validate:"required,min=1"`
}

type DeckWordUpdateRequest struct {
	Word    *string `json:"word,omitempty" validate:"omitempty,min=1,max=255"`
	Meaning *string `json:"meaning,omitempty" validate:"omitempty,min=1"`
}

type DeckWordResponse struct {
	ID        uuid.UUID `json:"id"`
	DeckID    uuid.UUID `json:"deck_id"`
	Word      string    `json:"word"`
	Meaning   string    `json:"meaning"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PaginatedDecksResponse struct {
	Decks      []DeckResponse `json:"decks"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	Total      int64          `json:"total"`
	TotalPages int            `json:"total_pages"`
}

type PaginatedDeckWordsResponse struct {
	Words      []DeckWordResponse `json:"words"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	Total      int64              `json:"total"`
	TotalPages int                `json:"total_pages"`
}