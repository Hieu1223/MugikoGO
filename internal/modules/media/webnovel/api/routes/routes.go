package routes

import (
	"database/sql"
	"net/http"

	"mugiko/internal/modules/media/webnovel/api/routes/handlers"
	"mugiko/internal/modules/media/webnovel/persistent"
	"mugiko/internal/modules/media/webnovel/service"
)

func RegisterWebNovelRoutes(mux *http.ServeMux, db *sql.DB) {
	novelRepo := persistent.NewNovelRepository(db)
	chapterRepo := persistent.NewChapterRepository(db)
	paragraphRepo := persistent.NewParagraphRepository(db)
	svc := service.NewNovelService(novelRepo, chapterRepo, paragraphRepo)
	handler := handlers.NewNovelHandler(svc)

	// Novel routes
	mux.HandleFunc("POST /api/v1/webnovel/novels", handler.AddNovel)
	mux.HandleFunc("GET /api/v1/webnovel/novels", handler.ListMyNovels)
	mux.HandleFunc("GET /api/v1/webnovel/novels/all", handler.ListAllNovels)
	mux.HandleFunc("GET /api/v1/webnovel/novels/{novel_id}", handler.GetNovel)
	mux.HandleFunc("POST /api/v1/webnovel/preview", handler.PreviewNovel)
	mux.HandleFunc("POST /api/v1/webnovel/search", handler.SearchNovels)

	// Chapter routes
	mux.HandleFunc("GET /api/v1/webnovel/chapters/{chapter_id}", handler.GetChapter)

}
