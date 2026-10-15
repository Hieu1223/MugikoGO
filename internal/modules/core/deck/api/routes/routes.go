package routes

import (
	"database/sql"
	"net/http"

	"mugiko/internal/modules/core/deck/api/routes/handlers"
	"mugiko/internal/modules/core/deck/service"
	"mugiko/internal/modules/core/deck/persistent"
)

func RegisterDeckRoutes(mux *http.ServeMux, db *sql.DB) {
	repo := persistent.NewDeckRepository(db)
	deckService := service.NewDeckService(repo)
	deckHandler := handlers.NewDeckHandler(deckService)

	if err := deckService.Initialize(); err != nil {
		panic("failed to initialize deck: " + err.Error())
	}

	// Deck routes
	mux.HandleFunc("POST /api/v1/decks", deckHandler.CreateDeck)
	mux.HandleFunc("GET /api/v1/decks", deckHandler.ListDecks)
	mux.HandleFunc("GET /api/v1/decks/{id}", deckHandler.GetDeck)
	mux.HandleFunc("PATCH /api/v1/decks/{id}", deckHandler.UpdateDeck)
	mux.HandleFunc("DELETE /api/v1/decks/{id}", deckHandler.DeleteDeck)

	// Word routes
	mux.HandleFunc("POST /api/v1/decks/{deck_id}/words", deckHandler.AddWord)
	mux.HandleFunc("GET /api/v1/decks/{deck_id}/words", deckHandler.ListWords)
	mux.HandleFunc("PATCH /api/v1/decks/{deck_id}/words/{word_id}", deckHandler.UpdateWord)
	mux.HandleFunc("DELETE /api/v1/decks/{deck_id}/words/{word_id}", deckHandler.DeleteWord)
}