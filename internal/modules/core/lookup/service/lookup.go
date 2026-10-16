package service

import (
	"mugiko/internal/modules/core/lookup/dto"
	"mugiko/internal/modules/core/lookup/persistent"
)

type LookupService struct {
	repo *persistent.LookupRepository
}

func NewLookupService(repo *persistent.LookupRepository) *LookupService {
	return &LookupService{repo: repo}
}

func (s *LookupService) LookupWord(word string) (*dto.DictionaryEntry, error) {
	entry := s.repo.LookupWord(word)
	if entry == nil {
		return nil, nil
	}
	return &dto.DictionaryEntry{
		Word: entry.Word,
		Readings: func() []dto.DictionaryReading {
			readings := make([]dto.DictionaryReading, len(entry.Readings))
			for i, r := range entry.Readings {
				readings[i] = dto.DictionaryReading{Kana: r.Kana, Meaning: r.Meaning}
			}
			return readings
		}(),
	}, nil
}

func (s *LookupService) PrefixSearch(prefix string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	entries := s.repo.PrefixSearch(prefix, limit)
	words := make([]string, len(entries))
	for i, e := range entries {
		words[i] = e.Word
	}
	return words, nil
}

func (s *LookupService) LookupKanji(kanji string) (*dto.KanjiEntry, error) {
	entry := s.repo.LookupKanji(kanji)
	if entry == nil {
		return nil, nil
	}
	return &dto.KanjiEntry{
		Kanji:      entry.Kanji,
		Reading:    entry.Reading,
		Meanings:   entry.Meanings,
		Strokes:    entry.Strokes,
		Radical:    entry.Radical,
		PenStrokes: entry.Penstrokes,
		Shape:      entry.Shape,
		Unicode:    entry.Unicode,
	}, nil
}
