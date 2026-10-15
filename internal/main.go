package internal

import (
	"database/sql"
	"mugiko/internal/common/middleware"
	log "mugiko/internal/modules/core/log"
	lookuproutes "mugiko/internal/modules/core/lookup/api/routes"
	deckroutes "mugiko/internal/modules/core/deck/api/routes"
	userroutes "mugiko/internal/modules/core/user/api/routes"
	userhandlers "mugiko/internal/modules/core/user/api/routes/handlers"
	user "mugiko/internal/modules/core/user/middleware"
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

	// Health endpoint
	mux.HandleFunc("GET /health", userhandlers.HealthCheck)

	http.ListenAndServe(":8080", stack(mux))
}
