package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mugiko/internal/common/api"
	"mugiko/internal/modules/core/lookup/dto"
	"mugiko/internal/modules/core/lookup/exception"
	"mugiko/internal/modules/core/lookup/service"
)

type LookupHandler struct {
	lookupService *service.LookupService
}

func NewLookupHandler(lookupService *service.LookupService) *LookupHandler {
	return &LookupHandler{lookupService: lookupService}
}

func (h *LookupHandler) LookupWord(w http.ResponseWriter, r *http.Request) {
	word := r.URL.Query().Get("word")
	if word == "" {
		api.WriteError(w, exception.MsgWordParameterRequired, http.StatusBadRequest)
		return
	}

	entry, err := h.lookupService.LookupWord(word)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entry == nil {
		api.WriteError(w, exception.MsgWordNotFound, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (h *LookupHandler) PrefixSearch(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		api.WriteError(w, exception.MsgPrefixParameterRequired, http.StatusBadRequest)
		return
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	results, err := h.lookupService.PrefixSearch(prefix, limit)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.PrefixSearchResponse{
		Results: results,
		Limit:   limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *LookupHandler) LookupKanji(w http.ResponseWriter, r *http.Request) {
	kanji := r.URL.Query().Get("kanji")
	if kanji == "" {
		api.WriteError(w, exception.MsgKanjiParameterRequired, http.StatusBadRequest)
		return
	}

	if len([]rune(kanji)) != 1 {
		api.WriteError(w, exception.MsgKanjiMustBeSingleCharacter, http.StatusBadRequest)
		return
	}

	entry, err := h.lookupService.LookupKanji(kanji)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entry == nil {
		api.WriteError(w, exception.MsgKanjiNotFound, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}
