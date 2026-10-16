package routes

import (
	"database/sql"
	"net/http"

	"mugiko/internal/modules/media/subtitle/api/routes/handlers"
	"mugiko/internal/modules/media/subtitle/persistent"
	"mugiko/internal/modules/media/subtitle/service"
)

func RegisterSubtitleRoutes(mux *http.ServeMux, db *sql.DB) {
	repo := persistent.NewSubtitleRepository(db)
	svc := service.NewSubtitleService(repo)
	handler := handlers.NewSubtitleHandler(svc)

	// Provider routes
	mux.HandleFunc("GET /api/v1/subtitle-providers", handler.ListProviders)
	mux.HandleFunc("POST /api/v1/subtitle-providers", handler.CreateProvider)

	// Subtitle routes
	mux.HandleFunc("GET /api/v1/subtitles", handler.SearchSubtitles)
	mux.HandleFunc("POST /api/v1/subtitles", handler.CreateSubtitle)
	mux.HandleFunc("GET /api/v1/subtitles/{subtitle_id}", handler.GetSubtitle)
	mux.HandleFunc("GET /api/v1/subtitles/{subtitle_id}/segments", handler.GetSubtitleSegments)

	// User history route
	mux.HandleFunc("GET /api/v1/users/me/subtitle-history", handler.GetUserHistory)

	// User settings routes
	mux.HandleFunc("GET /api/v1/users/me/subtitle-settings", handler.GetUserSettings)
	mux.HandleFunc("PUT /api/v1/users/me/subtitle-settings", handler.UpdateUserSettings)

	// User state routes
	mux.HandleFunc("GET /api/v1/subtitles/{subtitle_id}/state", handler.GetSubtitleUserState)
	mux.HandleFunc("PUT /api/v1/subtitles/{subtitle_id}/state", handler.UpdateSubtitleUserState)
}
