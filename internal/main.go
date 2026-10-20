package internal

import (
	"mugiko/internal/common/middleware"
	"mugiko/internal/common/persistent"
	deckroutes "mugiko/internal/modules/core/deck/api/routes"
	log "mugiko/internal/modules/core/log"
	lookuproutes "mugiko/internal/modules/core/lookup/api/routes"
	userroutes "mugiko/internal/modules/core/user/api/routes"
	userhandlers "mugiko/internal/modules/core/user/api/routes/handlers"
	user "mugiko/internal/modules/core/user/middleware"
	mangaroutes "mugiko/internal/modules/media/manga/api/routes"
	ocrroutes "mugiko/internal/modules/media/ocr/api/routes"
	subtitleroutes "mugiko/internal/modules/media/subtitle/api/routes"
	webnovelroutes "mugiko/internal/modules/media/webnovel/api/routes"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func Run() {
	db := persistent.GetDB()
	stack := middleware.Stack(
		log.Log,
		user.AuthContext,
	)
	mux := http.NewServeMux()

	// Register user routes
	userroutes.RegisterUserRoutes(mux)
	userroutes.RegisterAdminRoutes(mux)

	// Register lookup routes
	lookuproutes.RegisterLookupRoutes(mux, "asset/dictionary.bin", "asset/kanji_dictionary.bin")

	// Register deck routes
	deckroutes.RegisterDeckRoutes(mux, db)

	// Register subtitle routes
	subtitleroutes.RegisterSubtitleRoutes(mux, db)

	// Register OCR routes
	ocrroutes.RegisterOCRRoutes(mux, db)

	// Register manga routes
	mangaroutes.RegisterMangaRoutes(mux, db)

	// Register webnovel routes
	webnovelroutes.RegisterWebNovelRoutes(mux, db)

	// Health endpoint
	mux.HandleFunc("GET /health", userhandlers.HealthCheck)

	http.ListenAndServe(":"+os.Getenv("PORT"), stack(mux))
}
