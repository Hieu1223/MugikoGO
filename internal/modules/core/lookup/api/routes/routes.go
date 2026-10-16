package routes

import (
	"net/http"

	"mugiko/internal/modules/core/lookup/api/routes/handlers"
	"mugiko/internal/modules/core/lookup/persistent"
	"mugiko/internal/modules/core/lookup/service"
)

func RegisterLookupRoutes(mux *http.ServeMux, dictPath, kanjiDictPath string) {
	repo, err := persistent.NewLookupRepository(dictPath, kanjiDictPath)
	if err != nil {
		panic("failed to create lookup repository: " + err.Error())
	}
	lookupService := service.NewLookupService(repo)
	lookupHandler := handlers.NewLookupHandler(lookupService)

	mux.HandleFunc("GET /api/v1/lookup/word", lookupHandler.LookupWord)
	mux.HandleFunc("GET /api/v1/lookup/prefix", lookupHandler.PrefixSearch)
	mux.HandleFunc("GET /api/v1/lookup/kanji", lookupHandler.LookupKanji)
}
