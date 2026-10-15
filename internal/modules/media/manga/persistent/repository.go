package persistent

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"mugiko/internal/modules/media/manga/dto"
	"time"

	"github.com/google/uuid"
)

type MangaProvider struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	BaseURL     string
	Visibility  dto.ProviderVisibility
	IsBuiltin   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MangaSubscription struct {
	UserID     uuid.UUID
	ProviderID uuid.UUID
	CreatedAt  time.Time
}

type MangaHistory struct {
	UserID      uuid.UUID
	ProviderID  uuid.UUID
	MangaID     string
	ChapterID   string
	PageIndex   int
	Title       string
	CoverURL    sql.NullString
	LastReadAt  time.Time
}

type MangaUserSettings struct {
	UserID        uuid.UUID
	SchemaVersion int
	Data          []byte // JSON
	UpdatedAt     time.Time
}

type MangaRepository struct {
	db *sql.DB
}

func NewMangaRepository(db *sql.DB) *MangaRepository {
	return &MangaRepository{db: db}
}

// InitializeSchema creates all manga tables
func (r *MangaRepository) InitializeSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS manga_providers (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL,
			name TEXT NOT NULL,
			base_url TEXT NOT NULL,
			visibility TEXT NOT NULL DEFAULT 'private',
			is_builtin INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (owner_id) REFERENCES user_auths(user_id) ON DELETE CASCADE
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ix_manga_providers_owner_base_url ON manga_providers(owner_id, base_url)`,
		`CREATE INDEX IF NOT EXISTS ix_manga_providers_visibility ON manga_providers(visibility)`,
		`CREATE INDEX IF NOT EXISTS ix_manga_providers_owner_id ON manga_providers(owner_id)`,

		`CREATE TABLE IF NOT EXISTS manga_subscriptions (
			user_id TEXT NOT NULL,
			provider_id TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			PRIMARY KEY (user_id, provider_id),
			FOREIGN KEY (user_id) REFERENCES user_auths(user_id) ON DELETE CASCADE,
			FOREIGN KEY (provider_id) REFERENCES manga_providers(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS manga_history (
			user_id TEXT NOT NULL,
			provider_id TEXT NOT NULL,
			manga_id TEXT NOT NULL,
			chapter_id TEXT NOT NULL,
			page_index INTEGER NOT NULL DEFAULT 0,
			title TEXT NOT NULL,
			cover_url TEXT,
			last_read_at TEXT NOT NULL DEFAULT (datetime('now')),
			PRIMARY KEY (user_id, provider_id, manga_id),
			FOREIGN KEY (user_id) REFERENCES user_auths(user_id) ON DELETE CASCADE,
			FOREIGN KEY (provider_id) REFERENCES manga_providers(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS ix_manga_history_user_last_read ON manga_history(user_id, last_read_at)`,

		`CREATE TABLE IF NOT EXISTS manga_user_settings (
			user_id TEXT PRIMARY KEY,
			schema_version INTEGER NOT NULL DEFAULT 1,
			data TEXT NOT NULL DEFAULT '{}',
			updated_at TEXT NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (user_id) REFERENCES user_auths(user_id) ON DELETE CASCADE
		)`,
	}

	for _, q := range queries {
		if _, err := r.db.Exec(q); err != nil {
			return fmt.Errorf("schema init failed: %w", err)
		}
	}
	return nil
}

// Provider methods
func (r *MangaRepository) CreateProvider(ownerID uuid.UUID, name, baseURL string, visibility dto.ProviderVisibility, isBuiltin bool) (*MangaProvider, error) {
	id := uuid.New()
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`INSERT INTO manga_providers (id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id.String(), ownerID.String(), name, baseURL, string(visibility), isBuiltin, now.Format(time.RFC3339), now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	return &MangaProvider{
		ID: id, OwnerID: ownerID, Name: name, BaseURL: baseURL,
		Visibility: visibility, IsBuiltin: isBuiltin,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (r *MangaRepository) GetProvider(providerID uuid.UUID) (*MangaProvider, error) {
	row := r.db.QueryRow(`SELECT id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at FROM manga_providers WHERE id = ?`, providerID.String())
	var p MangaProvider
	var vis string
	var createdAt, updatedAt string
	if err := row.Scan(&p.ID, &p.OwnerID, &p.Name, &p.BaseURL, &vis, &p.IsBuiltin, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	p.Visibility = dto.ProviderVisibility(vis)
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &p, nil
}

func (r *MangaRepository) GetProviderByOwnerAndURL(ownerID uuid.UUID, baseURL string) (*MangaProvider, error) {
	row := r.db.QueryRow(`SELECT id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at FROM manga_providers WHERE owner_id = ? AND base_url = ?`, ownerID.String(), baseURL)
	var p MangaProvider
	var vis string
	var createdAt, updatedAt string
	if err := row.Scan(&p.ID, &p.OwnerID, &p.Name, &p.BaseURL, &vis, &p.IsBuiltin, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	p.Visibility = dto.ProviderVisibility(vis)
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &p, nil
}

func (r *MangaRepository) GetUserProviders(userID uuid.UUID) ([]*MangaProvider, error) {
	rows, err := r.db.Query(`SELECT id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at FROM manga_providers WHERE owner_id = ? ORDER BY created_at DESC`, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanProviders(rows)
}

func (r *MangaRepository) GetPublicProviders(page, pageSize int) ([]*MangaProvider, int, error) {
	// Count
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM manga_providers WHERE visibility = 'public' AND is_builtin = 0`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	// Query
	offset := (page - 1) * pageSize
	rows, err := r.db.Query(`SELECT id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at FROM manga_providers WHERE visibility = 'public' AND is_builtin = 0 ORDER BY created_at DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	providers, err := r.scanProviders(rows)
	return providers, total, err
}

func (r *MangaRepository) GetSubscribableProviders(userID uuid.UUID) ([]*MangaProvider, error) {
	rows, err := r.db.Query(`SELECT id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at FROM manga_providers WHERE visibility = 'public' OR owner_id = ? ORDER BY created_at DESC`, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanProviders(rows)
}

func (r *MangaRepository) UpdateProvider(provider *MangaProvider) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(`UPDATE manga_providers SET name = ?, base_url = ?, visibility = ?, updated_at = ? WHERE id = ?`,
		provider.Name, provider.BaseURL, string(provider.Visibility), now.Format(time.RFC3339), provider.ID.String())
	return err
}

func (r *MangaRepository) DeleteProvider(providerID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM manga_providers WHERE id = ?`, providerID.String())
	return err
}

// Subscription methods
func (r *MangaRepository) Subscribe(userID, providerID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(`INSERT OR REPLACE INTO manga_subscriptions (user_id, provider_id, created_at) VALUES (?, ?, ?)`,
		userID.String(), providerID.String(), now.Format(time.RFC3339))
	return err
}

func (r *MangaRepository) Unsubscribe(userID, providerID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM manga_subscriptions WHERE user_id = ? AND provider_id = ?`, userID.String(), providerID.String())
	return err
}

func (r *MangaRepository) GetSubscription(userID, providerID uuid.UUID) (*MangaSubscription, error) {
	row := r.db.QueryRow(`SELECT user_id, provider_id, created_at FROM manga_subscriptions WHERE user_id = ? AND provider_id = ?`, userID.String(), providerID.String())
	var s MangaSubscription
	var createdAt string
	if err := row.Scan(&s.UserID, &s.ProviderID, &createdAt); err != nil {
		return nil, err
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &s, nil
}

func (r *MangaRepository) GetUserSubscriptions(userID uuid.UUID) ([]*MangaSubscription, error) {
	rows, err := r.db.Query(`SELECT user_id, provider_id, created_at FROM manga_subscriptions WHERE user_id = ? ORDER BY created_at DESC`, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var subs []*MangaSubscription
	for rows.Next() {
		var s MangaSubscription
		var createdAt string
		if err := rows.Scan(&s.UserID, &s.ProviderID, &createdAt); err != nil {
			return nil, err
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		subs = append(subs, &s)
	}
	return subs, nil
}

func (r *MangaRepository) IsSubscribed(userID, providerID uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM manga_subscriptions WHERE user_id = ? AND provider_id = ?`, userID.String(), providerID.String()).Scan(&count)
	return count > 0, err
}

// History methods
func (r *MangaRepository) UpsertHistory(userID, providerID uuid.UUID, mangaID, chapterID string, pageIndex int, title, coverURL string) error {
	now := time.Now().UTC()
	cover := sql.NullString{String: coverURL, Valid: coverURL != ""}
	_, err := r.db.Exec(`INSERT INTO manga_history (user_id, provider_id, manga_id, chapter_id, page_index, title, cover_url, last_read_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, provider_id, manga_id) DO UPDATE SET
			chapter_id = excluded.chapter_id,
			page_index = excluded.page_index,
			title = excluded.title,
			cover_url = excluded.cover_url,
			last_read_at = excluded.last_read_at`,
		userID.String(), providerID.String(), mangaID, chapterID, pageIndex, title, cover, now.Format(time.RFC3339))
	return err
}

func (r *MangaRepository) GetUserHistory(userID uuid.UUID, page, pageSize int) ([]*MangaHistory, int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM manga_history WHERE user_id = ?`, userID.String()).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.Query(`SELECT user_id, provider_id, manga_id, chapter_id, page_index, title, cover_url, last_read_at FROM manga_history WHERE user_id = ? ORDER BY last_read_at DESC LIMIT ? OFFSET ?`, userID.String(), pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var history []*MangaHistory
	for rows.Next() {
		var h MangaHistory
		var lastRead string
		if err := rows.Scan(&h.UserID, &h.ProviderID, &h.MangaID, &h.ChapterID, &h.PageIndex, &h.Title, &h.CoverURL, &lastRead); err != nil {
			return nil, 0, err
		}
		h.LastReadAt, _ = time.Parse(time.RFC3339, lastRead)
		history = append(history, &h)
	}
	return history, total, nil
}

func (r *MangaRepository) DeleteHistory(userID, providerID uuid.UUID, mangaID string) error {
	_, err := r.db.Exec(`DELETE FROM manga_history WHERE user_id = ? AND provider_id = ? AND manga_id = ?`, userID.String(), providerID.String(), mangaID)
	return err
}

// Settings methods
func (r *MangaRepository) GetUserSettings(userID uuid.UUID) (*MangaUserSettings, error) {
	row := r.db.QueryRow(`SELECT user_id, schema_version, data, updated_at FROM manga_user_settings WHERE user_id = ?`, userID.String())
	var s MangaUserSettings
	var updatedAt string
	if err := row.Scan(&s.UserID, &s.SchemaVersion, &s.Data, &updatedAt); err != nil {
		return nil, err
	}
	s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &s, nil
}

func (r *MangaRepository) UpsertUserSettings(userID uuid.UUID, schemaVersion *int, data map[string]interface{}) (*MangaUserSettings, error) {
	now := time.Now().UTC()
	var jsonData []byte
	if data != nil {
		jsonData, _ = json.Marshal(data)
	} else {
		jsonData = []byte("{}")
	}
	sv := 1
	if schemaVersion != nil {
		sv = *schemaVersion
	}
	_, err := r.db.Exec(`INSERT INTO manga_user_settings (user_id, schema_version, data, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			schema_version = COALESCE(?, schema_version),
			data = COALESCE(?, data),
			updated_at = ?`,
		userID.String(), sv, string(jsonData), now.Format(time.RFC3339),
		sv, string(jsonData), now.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	return &MangaUserSettings{UserID: userID, SchemaVersion: sv, Data: jsonData, UpdatedAt: now}, nil
}

// Scanner helpers
func (r *MangaRepository) scanProviders(rows *sql.Rows) ([]*MangaProvider, error) {
	var providers []*MangaProvider
	for rows.Next() {
		var p MangaProvider
		var vis string
		var createdAt, updatedAt string
		if err := rows.Scan(&p.ID, &p.OwnerID, &p.Name, &p.BaseURL, &vis, &p.IsBuiltin, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		p.Visibility = dto.ProviderVisibility(vis)
		p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		providers = append(providers, &p)
	}
	return providers, nil
}