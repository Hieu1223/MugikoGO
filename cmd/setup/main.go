package main

import (
	"database/sql"
	"fmt"

	"mugiko/internal/common/persistent"
	deckpersistent "mugiko/internal/modules/core/deck/persistent"
	mangapersistent "mugiko/internal/modules/media/manga/persistent"
	subtitlepersistent "mugiko/internal/modules/media/subtitle/persistent"
	webnovelpersistent "mugiko/internal/modules/media/webnovel/persistent"
)

func main() {
	// Initialize database
	persistent.InitDB()
	defer persistent.CloseDB()

	db := persistent.GetDB()

	fmt.Println("Database setup complete!")
	fmt.Println("Tables created: user_auths, user_profiles")

	// Initialize all module schemas
	if err := initAllSchemas(db); err != nil {
		fmt.Printf("Schema initialization failed: %v\n", err)
		return
	}

	// Seed data
	if err := seedData(db); err != nil {
		fmt.Printf("Seed data failed: %v\n", err)
		return
	}

	// Create admin user if not exists
	createAdminIfNotExists()

	fmt.Println("Setup completed successfully!")
}

func initAllSchemas(db *sql.DB) error {
	fmt.Println("Initializing module schemas...")

	// Manga module
	mangaRepo := mangapersistent.NewMangaRepository(db)
	if err := mangaRepo.InitializeSchema(); err != nil {
		return fmt.Errorf("manga schema: %w", err)
	}
	fmt.Println("  ✓ manga module")

	// Deck module
	deckRepo := deckpersistent.NewDeckRepository(db)
	if err := deckRepo.InitializeSchema(); err != nil {
		return fmt.Errorf("deck schema: %w", err)
	}
	fmt.Println("  ✓ deck module")

	// Subtitle module
	subtitleRepo := subtitlepersistent.NewSubtitleRepository(db)
	if err := subtitleRepo.InitializeSchema(); err != nil {
		return fmt.Errorf("subtitle schema: %w", err)
	}
	fmt.Println("  ✓ subtitle module")

	// WebNovel module
	webnovelRepo := webnovelpersistent.NewNovelRepository(db)
	if err := webnovelRepo.InitializeSchema(); err != nil {
		return fmt.Errorf("webnovel schema: %w", err)
	}
	fmt.Println("  ✓ webnovel module")

	fmt.Println("All module schemas initialized!")
	return nil
}

func seedData(db *sql.DB) error {
	fmt.Println("Seeding initial data...")

	// Seed subtitle providers
	subtitleRepo := subtitlepersistent.NewSubtitleRepository(db)
	if err := seedSubtitleProviders(subtitleRepo); err != nil {
		return fmt.Errorf("subtitle providers seed: %w", err)
	}
	fmt.Println("  ✓ subtitle providers")

	fmt.Println("Seed data completed!")
	return nil
}

func seedSubtitleProviders(repo *subtitlepersistent.SubtitleRepository) error {
	// Check if providers already exist
	providers, err := repo.GetActiveProviders()
	if err != nil {
		return err
	}
	if len(providers) > 0 {
		fmt.Println("  subtitle providers already exist, skipping")
		return nil
	}

	// Seed default providers with their services
	defaultProviders := []struct {
		name     string
		kind     subtitlepersistent.ProviderKind
		baseURL  string
		services []subtitlepersistent.SubtitleProviderService
	}{
		{
			name:    "OpenSubtitles",
			kind:    subtitlepersistent.ProviderKindRemote,
			baseURL: "https://api.opensubtitles.com",
			services: []subtitlepersistent.SubtitleProviderService{
				{ServiceKey: "search", Name: "Search Subtitles"},
				{ServiceKey: "download", Name: "Download Subtitles"},
			},
		},
		{
			name:    "Subscene",
			kind:    subtitlepersistent.ProviderKindRemote,
			baseURL: "https://subscene.com",
			services: []subtitlepersistent.SubtitleProviderService{
				{ServiceKey: "search", Name: "Search Subtitles"},
				{ServiceKey: "download", Name: "Download Subtitles"},
			},
		},
		{
			name:    "YIFY Subtitles",
			kind:    subtitlepersistent.ProviderKindRemote,
			baseURL: "https://yifysubtitles.ch",
			services: []subtitlepersistent.SubtitleProviderService{
				{ServiceKey: "search", Name: "Search Subtitles"},
			},
		},
	}

	for _, p := range defaultProviders {
		provider, err := repo.CreateProvider(p.name, p.kind, &p.baseURL)
		if err != nil {
			return fmt.Errorf("failed to create provider %s: %w", p.name, err)
		}

		// Seed services for this provider
		if len(p.services) > 0 {
			// Convert to map format for UpsertServices
			svcMaps := make([]map[string]any, len(p.services))
			for i, svc := range p.services {
				svcMaps[i] = map[string]any{
					"service_key": svc.ServiceKey,
					"name":        svc.Name,
					"options":     nil,
				}
			}
			if err := repo.UpsertServices(provider.ID, svcMaps); err != nil {
				return fmt.Errorf("failed to create services for provider %s: %w", p.name, err)
			}
		}
	}

	return nil
}

func createAdminIfNotExists() {
	// Admin user creation would go here
	// For now, just print info
	fmt.Println("Admin user creation skipped - implement as needed")
}