package service

import (
	"github.com/google/uuid"
	"mugiko/internal/modules/core/deck/dto"
	"mugiko/internal/modules/core/deck/persistent"
)

type DeckService struct {
	repo *persistent.DeckRepository
}

func NewDeckService(repo *persistent.DeckRepository) *DeckService {
	return &DeckService{repo: repo}
}

func (s *DeckService) CreateDeck(ownerID uuid.UUID, req dto.DeckCreateRequest) (*dto.DeckResponse, error) {
	deck, err := s.repo.CreateDeck(ownerID, req.Name)
	if err != nil {
		return nil, err
	}
	return s.toDeckResponse(deck), nil
}

func (s *DeckService) GetDeck(deckID uuid.UUID) (*dto.DeckResponse, error) {
	deck, err := s.repo.GetDeck(deckID)
	if err != nil {
		return nil, err
	}
	if deck == nil {
		return nil, nil
	}
	return s.toDeckResponse(deck), nil
}

func (s *DeckService) ListDecks(ownerID uuid.UUID, page, pageSize int) (*dto.PaginatedDecksResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	decks, total, err := s.repo.ListDecks(ownerID, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DeckResponse, len(decks))
	for i, d := range decks {
		responses[i] = d
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.PaginatedDecksResponse{
		Decks:      responses,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *DeckService) UpdateDeck(deckID uuid.UUID, req dto.DeckUpdateRequest) (*dto.DeckResponse, error) {
	deck, err := s.repo.UpdateDeck(deckID, req.Name)
	if err != nil {
		return nil, err
	}
	if deck == nil {
		return nil, nil
	}
	return s.toDeckResponse(deck), nil
}

func (s *DeckService) DeleteDeck(deckID uuid.UUID) error {
	return s.repo.DeleteDeck(deckID)
}

func (s *DeckService) AddWordToDeck(deckID uuid.UUID, req dto.DeckWordCreateRequest) (*dto.DeckWordResponse, error) {
	word, err := s.repo.AddWordToDeck(deckID, req.Word, req.Meaning)
	if err != nil {
		return nil, err
	}
	return s.toDeckWordResponse(word), nil
}

func (s *DeckService) ListDeckWords(deckID uuid.UUID, page, pageSize int) (*dto.PaginatedDeckWordsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	words, total, err := s.repo.ListDeckWords(deckID, page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DeckWordResponse, len(words))
	for i, w := range words {
		responses[i] = w
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &dto.PaginatedDeckWordsResponse{
		Words:      responses,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *DeckService) UpdateDeckWord(wordID uuid.UUID, req dto.DeckWordUpdateRequest) (*dto.DeckWordResponse, error) {
	word, err := s.repo.UpdateDeckWord(wordID, req.Word, req.Meaning)
	if err != nil {
		return nil, err
	}
	if word == nil {
		return nil, nil
	}
	return s.toDeckWordResponse(word), nil
}

func (s *DeckService) DeleteDeckWord(wordID uuid.UUID) error {
	return s.repo.DeleteDeckWord(wordID)
}

func (s *DeckService) Initialize() error {
	return s.repo.InitializeSchema()
}

func (s *DeckService) toDeckResponse(d *dto.Deck) *dto.DeckResponse {
	return &dto.DeckResponse{
		ID:        d.ID,
		OwnerID:   d.OwnerID,
		Name:      d.Name,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func (s *DeckService) toDeckWordResponse(w *dto.DeckWord) *dto.DeckWordResponse {
	return &dto.DeckWordResponse{
		ID:        w.ID,
		DeckID:    w.DeckID,
		Word:      w.Word,
		Meaning:   w.Meaning,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}