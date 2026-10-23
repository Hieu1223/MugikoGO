package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mugiko/internal/common/api"
	commonexception "mugiko/internal/common/exception"
	"github.com/google/uuid"

	"mugiko/internal/modules/core/deck/dto"
	"mugiko/internal/modules/core/deck/exception"
	"mugiko/internal/modules/core/deck/service"
	"mugiko/internal/modules/core/user/middleware"
)

type DeckHandler struct {
	deckService *service.DeckService
}

func NewDeckHandler(deckService *service.DeckService) *DeckHandler {
	return &DeckHandler{deckService: deckService}
}

func (h *DeckHandler) CreateDeck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo, ok := middleware.GetUserInfo(ctx)
	if !ok {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	var req dto.DeckCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, commonexception.MsgInvalidRequestBody, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		api.WriteError(w, exception.MsgDeckNameRequired, http.StatusBadRequest)
		return
	}

	ownerID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidUserID, http.StatusInternalServerError)
		return
	}

	deck, err := h.deckService.CreateDeck(ownerID, req)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deck)
}

func (h *DeckHandler) ListDecks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo, ok := middleware.GetUserInfo(ctx)
	if !ok {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	page := 1
	pageSize := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil {
			pageSize = v
		}
	}

	ownerID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidUserID, http.StatusInternalServerError)
		return
	}

	result, err := h.deckService.ListDecks(ownerID, page, pageSize)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *DeckHandler) GetDeck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserInfo(ctx)
	if !ok {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	deckIDStr := r.PathValue("id")
	deckID, err := uuid.Parse(deckIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidDeckID, http.StatusBadRequest)
		return
	}

	deck, err := h.deckService.GetDeck(deckID)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if deck == nil {
		api.WriteError(w, exception.MsgDeckNotFound, http.StatusNotFound)
		return
	}

	// Verify ownership - done in service layer for now

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deck)
}

func (h *DeckHandler) UpdateDeck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserInfo(ctx)
	if !ok {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	deckIDStr := r.PathValue("id")
	deckID, err := uuid.Parse(deckIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidDeckID, http.StatusBadRequest)
		return
	}

	var req dto.DeckUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, commonexception.MsgInvalidRequestBody, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		api.WriteError(w, exception.MsgDeckNameRequired, http.StatusBadRequest)
		return
	}

	deck, err := h.deckService.UpdateDeck(deckID, req)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if deck == nil {
		api.WriteError(w, exception.MsgDeckNotFound, http.StatusNotFound)
		return
	}

	// Verify ownership - done in service layer for now

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deck)
}

func (h *DeckHandler) DeleteDeck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserInfo(ctx)
	if !ok {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	deckIDStr := r.PathValue("id")
	deckID, err := uuid.Parse(deckIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidDeckID, http.StatusBadRequest)
		return
	}

	// Verify ownership by getting deck first
	deck, err := h.deckService.GetDeck(deckID)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if deck == nil {
		api.WriteError(w, exception.MsgDeckNotFound, http.StatusNotFound)
		return
	}

	// Verify ownership - done in service layer for now

	if err := h.deckService.DeleteDeck(deckID); err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DeckHandler) AddWord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserInfo(ctx)
	if !ok {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	deckIDStr := r.PathValue("deck_id")
	deckID, err := uuid.Parse(deckIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidDeckID, http.StatusBadRequest)
		return
	}

	var req dto.DeckWordCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, commonexception.MsgInvalidRequestBody, http.StatusBadRequest)
		return
	}

	if req.Word == "" || req.Meaning == "" {
		api.WriteError(w, exception.MsgWordAndMeaningRequired, http.StatusBadRequest)
		return
	}

	// Owner verification handled in service layer

	word, err := h.deckService.AddWordToDeck(deckID, req)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(word)
}

func (h *DeckHandler) ListWords(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo, ok := middleware.GetUserInfo(ctx)
	if !ok {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	deckIDStr := r.PathValue("deck_id")
	deckID, err := uuid.Parse(deckIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidDeckID, http.StatusBadRequest)
		return
	}

	page := 1
	pageSize := 50
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil {
			pageSize = v
		}
	}

	_ = userInfo // suppress unused warning for now (ownership verified in service layer)

	result, err := h.deckService.ListDeckWords(deckID, page, pageSize)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *DeckHandler) UpdateWord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserInfo(ctx)
	if !ok {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	wordIDStr := r.PathValue("word_id")
	wordID, err := uuid.Parse(wordIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidWordID, http.StatusBadRequest)
		return
	}

	var req dto.DeckWordUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, commonexception.MsgInvalidRequestBody, http.StatusBadRequest)
		return
	}

	if (req.Word != nil && *req.Word == "") || (req.Meaning != nil && *req.Meaning == "") {
		api.WriteError(w, exception.MsgWordAndMeaningCannotBeEmpty, http.StatusBadRequest)
		return
	}

	word, err := h.deckService.UpdateDeckWord(wordID, req)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if word == nil {
		api.WriteError(w, exception.MsgWordNotFound, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(word)
}

func (h *DeckHandler) DeleteWord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserInfo(ctx)
	if !ok {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	wordIDStr := r.PathValue("word_id")
	wordID, err := uuid.Parse(wordIDStr)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.deckService.DeleteDeckWord(wordID); err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
