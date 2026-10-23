package service

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"mugiko/internal/modules/media/ocr/dto"
	"mugiko/internal/modules/media/ocr/exception"
	"mugiko/internal/modules/media/ocr/persistent"
)

var ErrInvalidImageURL = errors.New(exception.MsgInvalidImageURL)

type OCRService struct {
	repo *persistent.OCRRepository
}

func NewOCRService(repo *persistent.OCRRepository) *OCRService {
	return &OCRService{repo: repo}
}

// CreateOCR creates a queued OCR job. It does not perform OCR.
func (s *OCRService) CreateOCR(userID uuid.UUID, req dto.OCRCreate) (*dto.OCRResponse, error) {
	imageURL, err := normalizeImageURL(req.ImageURL)
	if err != nil {
		return nil, err
	}

	result, err := s.repo.CreateOCR(imageURL, userID)
	if err != nil {
		return nil, err
	}

	return toResponse(result)
}

func toResponse(result *persistent.OCRResult) (*dto.OCRResponse, error) {
	response := &dto.OCRResponse{
		ID:        result.ID,
		ImageURL:  result.ImageURL,
		Status:    result.Status,
		Error:     result.Error,
		CreatedBy: result.CreatedBy,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	}

	if result.OCRData != nil {
		var data dto.PaddleOCRData
		if err := json.Unmarshal([]byte(*result.OCRData), &data); err != nil {
			return nil, err
		}
		response.OCRData = &data
	}

	return response, nil
}

func normalizeImageURL(raw string) (string, error) {
	imageURL := strings.TrimSpace(raw)
	if imageURL == "" {
		return "", ErrInvalidImageURL
	}

	parsed, err := url.ParseRequestURI(imageURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", ErrInvalidImageURL
	}

	return imageURL, nil
}
