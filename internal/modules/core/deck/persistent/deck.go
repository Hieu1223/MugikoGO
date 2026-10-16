package persistent

import (
	"database/sql"

	"mugiko/internal/modules/core/deck/dto"

	"github.com/google/uuid"
)

type DeckRepository struct {
	db *sql.DB
}

func NewDeckRepository(db *sql.DB) *DeckRepository {
	return &DeckRepository{db: db}
}

func (r *DeckRepository) CreateDeck(ownerID uuid.UUID, name string) (*dto.Deck, error) {
	deckID := uuid.New()

	_, err := r.db.Exec(`
		INSERT INTO decks (id, owner_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC')
	`, deckID.String(), ownerID.String(), name)
	if err != nil {
		return nil, err
	}

	return r.GetDeck(deckID)
}

func (r *DeckRepository) GetDeck(deckID uuid.UUID) (*dto.Deck, error) {
	var deck dto.Deck
	err := r.db.QueryRow(`
		SELECT id, owner_id, name, created_at, updated_at
		FROM decks
		WHERE id = $1
	`, deckID.String()).Scan(&deck.ID, &deck.OwnerID, &deck.Name, &deck.CreatedAt, &deck.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &deck, nil
}

func (r *DeckRepository) ListDecks(ownerID uuid.UUID, page, pageSize int) ([]dto.DeckResponse, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM decks WHERE owner_id = $1`, ownerID.String()).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
		SELECT id, owner_id, name, created_at, updated_at
		FROM decks
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, ownerID.String(), pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var decks []dto.DeckResponse
	for rows.Next() {
		var deck dto.DeckResponse
		if err := rows.Scan(&deck.ID, &deck.OwnerID, &deck.Name, &deck.CreatedAt, &deck.UpdatedAt); err != nil {
			return nil, 0, err
		}
		decks = append(decks, deck)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return decks, total, nil
}

func (r *DeckRepository) UpdateDeck(deckID uuid.UUID, name string) (*dto.Deck, error) {
	_, err := r.db.Exec(`
		UPDATE decks SET name = $1, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $2
	`, name, deckID.String())
	if err != nil {
		return nil, err
	}

	return r.GetDeck(deckID)
}

func (r *DeckRepository) DeleteDeck(deckID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM decks WHERE id = $1`, deckID.String())
	return err
}

func (r *DeckRepository) AddWordToDeck(deckID uuid.UUID, word, meaning string) (*dto.DeckWord, error) {
	wordID := uuid.New()
	_, err := r.db.Exec(`
		INSERT INTO deck_words (id, deck_id, word, meaning, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC')
	`, wordID.String(), deckID.String(), word, meaning)
	if err != nil {
		return nil, err
	}

	return r.GetDeckWord(wordID)
}

func (r *DeckRepository) GetDeckWord(wordID uuid.UUID) (*dto.DeckWord, error) {
	var dw dto.DeckWord
	err := r.db.QueryRow(`
		SELECT id, deck_id, word, meaning, created_at, updated_at
		FROM deck_words
		WHERE id = $1
	`, wordID.String()).Scan(&dw.ID, &dw.DeckID, &dw.Word, &dw.Meaning, &dw.CreatedAt, &dw.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &dw, nil
}

func (r *DeckRepository) ListDeckWords(deckID uuid.UUID, page, pageSize int) ([]dto.DeckWordResponse, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM deck_words WHERE deck_id = $1`, deckID.String()).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
		SELECT id, deck_id, word, meaning, created_at, updated_at
		FROM deck_words
		WHERE deck_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, deckID.String(), pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var words []dto.DeckWordResponse
	for rows.Next() {
		var w dto.DeckWordResponse
		if err := rows.Scan(&w.ID, &w.DeckID, &w.Word, &w.Meaning, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, 0, err
		}
		words = append(words, w)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return words, total, nil
}

func (r *DeckRepository) UpdateDeckWord(wordID uuid.UUID, word, meaning *string) (*dto.DeckWord, error) {
	var current dto.DeckWord
	err := r.db.QueryRow(`
		SELECT id, deck_id, word, meaning, created_at, updated_at
		FROM deck_words WHERE id = $1
	`, wordID.String()).Scan(&current.ID, &current.DeckID, &current.Word, &current.Meaning, &current.CreatedAt, &current.UpdatedAt)
	if err != nil {
		return nil, err
	}

	newWord := current.Word
	if word != nil {
		newWord = *word
	}
	newMeaning := current.Meaning
	if meaning != nil {
		newMeaning = *meaning
	}

	_, err = r.db.Exec(`
		UPDATE deck_words SET word = $1, meaning = $2, updated_at = NOW() AT TIME ZONE 'UTC'
		WHERE id = $3
	`, newWord, newMeaning, wordID.String())
	if err != nil {
		return nil, err
	}

	return r.GetDeckWord(wordID)
}

func (r *DeckRepository) DeleteDeckWord(wordID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM deck_words WHERE id = $1`, wordID.String())
	return err
}
