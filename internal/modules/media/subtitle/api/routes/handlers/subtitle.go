package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"mugiko/internal/common/api"
	commonexception "mugiko/internal/common/exception"
	"github.com/google/uuid"

	"mugiko/internal/modules/core/user/middleware"
	"mugiko/internal/modules/media/subtitle/dto"
	"mugiko/internal/modules/media/subtitle/exception"
	"mugiko/internal/modules/media/subtitle/service"
)

type SubtitleHandler struct {
	svc *service.SubtitleService
}

func NewSubtitleHandler(svc *service.SubtitleService) *SubtitleHandler {
	return &SubtitleHandler{svc: svc}
}

// Helper to get user ID from context
func getUserID(r *http.Request) *uuid.UUID {
	if userInfo, ok := middleware.GetUserInfo(r.Context()); ok {
		if userID, err := uuid.Parse(userInfo.UserID); err == nil {
			return &userID
		}
	}
	return nil
}

func (h *SubtitleHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.WriteError(w, commonexception.MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	providers, err := h.svc.ListProviders()
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

func (h *SubtitleHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.WriteError(w, commonexception.MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Check admin role
	userInfo, ok := middleware.GetUserInfo(r.Context())
	if !ok || userInfo.Role != "admin" {
		api.WriteError(w, commonexception.MsgAdminRequired, http.StatusForbidden)
		return
	}

	var req dto.SubtitleProviderCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, commonexception.MsgInvalidRequestBody, http.StatusBadRequest)
		return
	}

	provider, err := h.svc.CreateProvider(req)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(provider)
}

func (h *SubtitleHandler) CreateSubtitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		api.WriteError(w, commonexception.MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	if userID == nil {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	var req dto.SubtitleCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, commonexception.MsgInvalidRequestBody, http.StatusBadRequest)
		return
	}

	subtitle, err := h.svc.CreateSubtitle(*userID, req)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subtitle)
}

func (h *SubtitleHandler) SearchSubtitles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.WriteError(w, commonexception.MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	videoURL := r.URL.Query().Get("video_url")
	if videoURL == "" {
		api.WriteError(w, exception.MsgVideoURLQueryParameterRequired, http.StatusBadRequest)
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	subtitles, err := h.svc.SearchSubtitles(videoURL, limit, offset)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subtitles)
}

func (h *SubtitleHandler) GetSubtitle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.WriteError(w, commonexception.MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	subtitleIDStr := r.PathValue("subtitle_id")
	subtitleID, err := uuid.Parse(subtitleIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidSubtitleID, http.StatusBadRequest)
		return
	}

	userID := getUserID(r)

	subtitle, err := h.svc.GetSubtitle(subtitleID, userID)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subtitle)
}

func (h *SubtitleHandler) GetSubtitleSegments(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.WriteError(w, commonexception.MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	subtitleIDStr := r.PathValue("subtitle_id")
	subtitleID, err := uuid.Parse(subtitleIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidSubtitleID, http.StatusBadRequest)
		return
	}

	limit := 100
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	segments, err := h.svc.GetSubtitleSegments(subtitleID, limit, offset)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(segments)
}

func (h *SubtitleHandler) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.WriteError(w, commonexception.MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	if userID == nil {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	settings, err := h.svc.GetUserSettings(*userID)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (h *SubtitleHandler) UpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		api.WriteError(w, commonexception.MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	if userID == nil {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	var req dto.UserSettingsUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, commonexception.MsgInvalidRequestBody, http.StatusBadRequest)
		return
	}

	settings, err := h.svc.UpdateUserSettings(*userID, req)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (h *SubtitleHandler) GetSubtitleUserState(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.WriteError(w, commonexception.MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	if userID == nil {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	subtitleIDStr := r.PathValue("subtitle_id")
	subtitleID, err := uuid.Parse(subtitleIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidSubtitleID, http.StatusBadRequest)
		return
	}

	state, err := h.svc.GetSubtitleUserState(*userID, subtitleID)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (h *SubtitleHandler) UpdateSubtitleUserState(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		api.WriteError(w, commonexception.MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	if userID == nil {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	subtitleIDStr := r.PathValue("subtitle_id")
	subtitleID, err := uuid.Parse(subtitleIDStr)
	if err != nil {
		api.WriteError(w, exception.MsgInvalidSubtitleID, http.StatusBadRequest)
		return
	}

	var req dto.SubtitleUserStateUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, commonexception.MsgInvalidRequestBody, http.StatusBadRequest)
		return
	}

	state, err := h.svc.UpdateSubtitleUserState(*userID, subtitleID, req)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (h *SubtitleHandler) GetUserHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.WriteError(w, commonexception.MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	if userID == nil {
		api.WriteError(w, commonexception.MsgUnauthorized, http.StatusUnauthorized)
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	history, err := h.svc.GetUserHistory(*userID, limit, offset)
	if err != nil {
		api.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
