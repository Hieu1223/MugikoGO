package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mugiko/internal/modules/core/user/middleware"
	"mugiko/internal/modules/media/webnovel/dto"
	"mugiko/internal/modules/media/webnovel/service"

	"github.com/google/uuid"
)

type NovelHandler struct {
	svc *service.NovelService
}

func NewNovelHandler(svc *service.NovelService) *NovelHandler {
	return &NovelHandler{svc: svc}
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

// Novel handlers
func (h *NovelHandler) AddNovel(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req dto.AddNovelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	resp, err := h.svc.AddNovel(r, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *NovelHandler) ListMyNovels(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	page := parseQueryInt(r, "page", 1)
	pageSize := parseQueryInt(r, "page_size", 20)

	resp, err := h.svc.ListNovelsByUser(r, page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *NovelHandler) ListAllNovels(w http.ResponseWriter, r *http.Request) {
	page := parseQueryInt(r, "page", 1)
	pageSize := parseQueryInt(r, "page_size", 20)

	resp, err := h.svc.ListAllNovels(page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *NovelHandler) GetNovel(w http.ResponseWriter, r *http.Request) {
	novelID, err := parseUUID(r, "novel_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid novel ID"})
		return
	}

	resp, err := h.svc.GetNovel(novelID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "novel not found"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *NovelHandler) PreviewNovel(w http.ResponseWriter, r *http.Request) {
	var req dto.PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	resp, err := h.svc.Preview(r, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *NovelHandler) SearchNovels(w http.ResponseWriter, r *http.Request) {
	var req dto.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	resp, err := h.svc.Search(req.Query, req.Limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Chapter handlers
func (h *NovelHandler) GetChapter(w http.ResponseWriter, r *http.Request) {
	chapterID, err := parseUUID(r, "chapter_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid chapter ID"})
		return
	}

	limit := parseQueryInt(r, "limit", 0)
	offset := parseQueryInt(r, "offset", 0)

	var resp *dto.ChapterResponse
	if limit > 0 {
		resp, err = h.svc.GetChapterPaginated(chapterID, limit, offset)
	} else {
		resp, err = h.svc.GetChapter(chapterID)
	}

	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "chapter not found"})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Sync handler (not implemented)
func (h *NovelHandler) SyncNovel(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	novelID, err := parseUUID(r, "novel_id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid novel ID"})
		return
	}

	var req dto.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	_, err = h.svc.Sync(novelID, req.Force)
	if err != nil {
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "sync started"})
}