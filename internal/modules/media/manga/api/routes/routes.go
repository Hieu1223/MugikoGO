package routes

import (
	"database/sql"
	"net/http"

	"mugiko/internal/modules/media/manga/api/routes/handlers"
	"mugiko/internal/modules/media/manga/persistent"
	"mugiko/internal/modules/media/manga/service"
)

func RegisterMangaRoutes(mux *http.ServeMux, db *sql.DB) {
	repo := persistent.NewMangaRepository(db)
	svc := service.NewMangaService(repo)
	handler := handlers.NewMangaHandler(svc)

	// Provider routes
	mux.HandleFunc("GET /api/v1/manga/providers", handler.ListProviders)
	mux.HandleFunc("POST /api/v1/manga/providers", handler.CreateProvider)
	mux.HandleFunc("GET /api/v1/manga/providers/{provider_id}", handler.GetProvider)
	mux.HandleFunc("PUT /api/v1/manga/providers/{provider_id}", handler.UpdateProvider)
	mux.HandleFunc("DELETE /api/v1/manga/providers/{provider_id}", handler.DeleteProvider)

	// Public provider routes
	mux.HandleFunc("GET /api/v1/manga/providers/public", handler.ListPublicProviders)
	mux.HandleFunc("GET /api/v1/manga/providers/subscribable", handler.ListSubscribableProviders)

	// Subscription routes
	mux.HandleFunc("POST /api/v1/manga/providers/{provider_id}/subscribe", handler.Subscribe)
	mux.HandleFunc("DELETE /api/v1/manga/providers/{provider_id}/subscribe", handler.Unsubscribe)
	mux.HandleFunc("GET /api/v1/manga/subscriptions", handler.ListSubscriptions)
	mux.HandleFunc("GET /api/v1/manga/providers/{provider_id}/subscribed", handler.CheckSubscription)

	// History routes
	mux.HandleFunc("POST /api/v1/manga/history", handler.UpsertHistory)
	mux.HandleFunc("GET /api/v1/manga/history", handler.GetHistory)
	mux.HandleFunc("DELETE /api/v1/manga/providers/{provider_id}/history/{manga_id}", handler.DeleteHistory)

	// Settings routes
	mux.HandleFunc("GET /api/v1/users/me/manga-settings", handler.GetUserSettings)
	mux.HandleFunc("PUT /api/v1/users/me/manga-settings", handler.UpdateUserSettings)
}
