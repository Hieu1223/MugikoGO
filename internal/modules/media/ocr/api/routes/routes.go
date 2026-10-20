package routes

import (
	"database/sql"
	"net/http"

	"mugiko/internal/modules/media/ocr/api/routes/handlers"
	"mugiko/internal/modules/media/ocr/persistent"
	"mugiko/internal/modules/media/ocr/service"
)

func RegisterOCRRoutes(mux *http.ServeMux, db *sql.DB) {
	repo := persistent.NewOCRRepository(db)
	svc := service.NewOCRService(repo)
	handler := handlers.NewOCRHandler(svc)

	mux.HandleFunc("POST /api/v1/ocr", handler.CreateOCR)
}
