package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mugiko/internal/modules/core/user/middleware"
	"mugiko/internal/modules/media/manga/dto"
	"mugiko/internal/modules/media/manga/service"

	"github.com/google/uuid"
)

type MangaHandler struct {
	svc *service.MangaService
}

func NewMangaHandler(svc *service.MangaService) *MangaHandler {
	return &MangaHandler{svc: svc}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func parseUUID(r *http.Request, name string) (uuid.UUID, error) {
	val := r.PathValue(name)
	return uuid.Parse(val)
}

func parseQueryInt(r *http.Request, name string, defaultVal int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

// Provider handlers
func (h *MangaHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	providers, err := h.svc.GetUserProviders(r, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	result := make([]*dto.MangaProviderDTO, len(providers))
	for i, p := range providers {
		result[i] = h.svc.ToProviderDTO(p)
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *MangaHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req dto.CreateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	provider, err := h.svc.CreateProvider(r, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, provider)
}

func (h *MangaHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	providerID, err := parseUUID(r, "provider_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid provider ID"})
		return
	}

	provider, err := h.svc.GetProvider(providerID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "provider not found"})
		return
	}

	if provider.OwnerID != userID && user.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	writeJSON(w, http.StatusOK, provider)
}

func (h *MangaHandler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	providerID, err := parseUUID(r, "provider_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid provider ID"})
		return
	}

	var req dto.UpdateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	provider, err := h.svc.UpdateProvider(r, providerID, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, provider)
}

func (h *MangaHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	providerID, err := parseUUID(r, "provider_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid provider ID"})
		return
	}

	provider, err := h.svc.GetProvider(providerID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "provider not found"})
		return
	}

	if provider.OwnerID != userID && user.Role != "admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	if err := h.svc.DeleteProvider(r, providerID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *MangaHandler) ListPublicProviders(w http.ResponseWriter, r *http.Request) {
	page := parseQueryInt(r, "page", 1)
	pageSize := parseQueryInt(r, "page_size", 20)

	providers, total, err := h.svc.ListPublicProviders(page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":     providers,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *MangaHandler) ListSubscribableProviders(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	providers, err := h.svc.GetSubscribableProviders(r, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, providers)
}

// Subscription handlers
func (h *MangaHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	providerID, err := parseUUID(r, "provider_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid provider ID"})
		return
	}

	if err := h.svc.Subscribe(r, providerID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *MangaHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	providerID, err := parseUUID(r, "provider_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid provider ID"})
		return
	}

	if err := h.svc.Unsubscribe(r, providerID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *MangaHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	subs, err := h.svc.ListSubscriptions(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, subs)
}

func (h *MangaHandler) CheckSubscription(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	providerID, err := parseUUID(r, "provider_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid provider ID"})
		return
	}

	subscribed, err := h.svc.IsSubscribed(r, providerID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"subscribed": subscribed})
}

// History handlers
func (h *MangaHandler) UpsertHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req dto.UpsertHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if err := h.svc.UpsertHistory(r, req); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *MangaHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	page := parseQueryInt(r, "page", 1)
	pageSize := parseQueryInt(r, "page_size", 20)

	history, total, err := h.svc.GetUserHistory(r, page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":     history,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *MangaHandler) DeleteHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	providerID, err := parseUUID(r, "provider_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid provider ID"})
		return
	}

	mangaID := r.PathValue("manga_id")
	if mangaID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing manga_id"})
		return
	}

	if err := h.svc.DeleteHistory(r, providerID, mangaID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

// Settings handlers
func (h *MangaHandler) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	settings, err := h.svc.GetUserSettings(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *MangaHandler) UpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req dto.UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	settings, err := h.svc.UpdateUserSettings(r, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, settings)
}
