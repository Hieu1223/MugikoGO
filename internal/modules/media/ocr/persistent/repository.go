package persistent

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type OCRRepository struct {
	db *sql.DB
}

func NewOCRRepository(db *sql.DB) *OCRRepository {
	return &OCRRepository{db: db}
}

// CreateOCR inserts a queued OCR job. The external worker is responsible for
// filling ocr_data and advancing the status later.
func (r *OCRRepository) CreateOCR(imageURL string, createdBy uuid.UUID) (*OCRResult, error) {
	id := uuid.New()
	now := time.Now()

	_, err := r.db.Exec(`
		INSERT INTO ocr_results (id, image_url, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id.String(), imageURL, OCRStatusQueued, createdBy.String(), now, now)
	if err != nil {
		return nil, err
	}

	return &OCRResult{
		ID:        id,
		ImageURL:  imageURL,
		Status:    OCRStatusQueued,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
