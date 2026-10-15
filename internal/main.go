package internal

import (
	"database/sql"
	"fmt"
	"mugiko/internal/common/middleware"
	deckroutes "mugiko/internal/modules/core/deck/api/routes"
	log "mugiko/internal/modules/core/log"
	lookuproutes "mugiko/internal/modules/core/lookup/api/routes"
	mangaroutes "mugiko/internal/modules/media/manga/api/routes"
	subtitleroutes "mugiko/internal/modules/media/subtitle/api/routes"
	userroutes "mugiko/internal/modules/core/user/api/routes"
	userhandlers "mugiko/internal/modules/core/user/api/routes/handlers"
	user "mugiko/internal/modules/core/user/middleware"
	webnovelroutes "mugiko/internal/modules/media/webnovel/api/routes"
	"net/http"

	_ "modernc.org/sqlite"
)

func Run() {
	// Open database
	db, err := sql.Open("sqlite", "./mugiko.db")

	if err != nil {
		panic("failed to open database: " + err.Error())
	}
	defer db.Close()

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		panic("failed to WAL: " + err.Error())
	}

	// Recommended: Set synchronous to NORMAL for better WAL performance
	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		panic("failed to Sync Normal: " + err.Error())
	}

	fmt.Println("WAL mode successfully enabled!")

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

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

	// Register manga routes
	mangaroutes.RegisterMangaRoutes(mux, db)

	// Register webnovel routes
	webnovelroutes.RegisterWebNovelRoutes(mux, db)

	// Health endpoint
	mux.HandleFunc("GET /health", userhandlers.HealthCheck)

	http.ListenAndServe(":8080", stack(mux))
}
