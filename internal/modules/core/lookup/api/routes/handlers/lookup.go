package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mugiko/internal/modules/core/lookup/dto"
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
		writeError(w, "Word parameter is required", http.StatusBadRequest)
		return
	}

	entry, err := h.lookupService.LookupWord(word)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entry == nil {
		writeError(w, "Word not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (h *LookupHandler) PrefixSearch(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		writeError(w, "Prefix parameter is required", http.StatusBadRequest)
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
		writeError(w, err.Error(), http.StatusInternalServerError)
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
		writeError(w, "Kanji parameter is required", http.StatusBadRequest)
		return
	}

	if len([]rune(kanji)) != 1 {
		writeError(w, "Kanji must be a single character", http.StatusBadRequest)
		return
	}

	entry, err := h.lookupService.LookupKanji(kanji)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entry == nil {
		writeError(w, "Kanji not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error_code": "LOOKUP_ERROR",
		"detail":     message,
	})
}