package dto

import (
	"time"

	"github.com/google/uuid"
)

// OCRCreate is the request body for creating an OCR job.
type OCRCreate struct {
	ImageURL string `json:"image_url"`
}

// PaddleOCRData is the PaddleOCR result shape stored in ocr_results.ocr_data.
// Scores are intentionally omitted; only the text and polygons are retained.
type PaddleOCRData struct {
	DTPolys  [][][]float64 `json:"dt_polys"`
	RecTexts []string      `json:"rec_texts"`
	RecPolys [][][]float64 `json:"rec_polys"`
}

// OCRResponse is the API representation of an OCR job.
type OCRResponse struct {
	ID        uuid.UUID      `json:"id"`
	ImageURL  string         `json:"image_url"`
	Status    string         `json:"status"`
	OCRData   *PaddleOCRData `json:"ocr_data"`
	Error     *string        `json:"error"`
	CreatedBy uuid.UUID      `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
