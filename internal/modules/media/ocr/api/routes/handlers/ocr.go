package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"mugiko/internal/common/api"
	commonexception "mugiko/internal/common/exception"
	"mugiko/internal/modules/core/user/middleware"
	"mugiko/internal/modules/media/ocr/dto"
	"mugiko/internal/modules/media/ocr/exception"
	"mugiko/internal/modules/media/ocr/service"
)

type OCRHandler struct {
	svc *service.OCRService
}

func NewOCRHandler(svc *service.OCRService) *OCRHandler {
	return &OCRHandler{svc: svc}
}

func (h *OCRHandler) CreateOCR(w http.ResponseWriter, r *http.Request) {
	userInfo, ok := middleware.GetUserInfo(r.Context())
	if !ok {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	var req dto.OCRCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, commonexception.MsgInvalidRequestBody, http.StatusBadRequest)
		return
	}

	result, err := h.svc.CreateOCR(userID, req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidImageURL) {
			api.WriteError(w, err.Error(), http.StatusBadRequest)
			return
		}

		api.WriteError(w, exception.MsgFailedToCreateOCRJob, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}
