package persistent

import (
	"database/sql"
	"encoding/json"
	"mugiko/internal/modules/media/manga/dto"
	"time"

	"github.com/google/uuid"
)

func NewMangaRepository(db *sql.DB) *MangaRepository {
	return &MangaRepository{db: db}
}

// Provider methods
func (r *MangaRepository) CreateProvider(ownerID uuid.UUID, name, baseURL string, visibility dto.ProviderVisibility, isBuiltin bool) (*MangaProvider, error) {
	id := uuid.New()
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`INSERT INTO manga_providers (id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
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
	row := r.db.QueryRow(`SELECT id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at FROM manga_providers WHERE id = $1`, providerID.String())
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
	row := r.db.QueryRow(`SELECT id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at FROM manga_providers WHERE owner_id = $1 AND base_url = $2`, ownerID.String(), baseURL)
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
	rows, err := r.db.Query(`SELECT id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at FROM manga_providers WHERE owner_id = $1 ORDER BY created_at DESC`, userID.String())
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
	rows, err := r.db.Query(`SELECT id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at FROM manga_providers WHERE visibility = 'public' AND is_builtin = 0 ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	providers, err := r.scanProviders(rows)
	return providers, total, err
}

func (r *MangaRepository) GetSubscribableProviders(userID uuid.UUID) ([]*MangaProvider, error) {
	rows, err := r.db.Query(`SELECT id, owner_id, name, base_url, visibility, is_builtin, created_at, updated_at FROM manga_providers WHERE visibility = 'public' OR owner_id = $1 ORDER BY created_at DESC`, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanProviders(rows)
}

func (r *MangaRepository) UpdateProvider(provider *MangaProvider) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(`UPDATE manga_providers SET name = $1, base_url = $2, visibility = $3, updated_at = $4 WHERE id = $5`,
		provider.Name, provider.BaseURL, string(provider.Visibility), now.Format(time.RFC3339), provider.ID.String())
	return err
}

func (r *MangaRepository) DeleteProvider(providerID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM manga_providers WHERE id = $1`, providerID.String())
	return err
}

// Subscription methods
func (r *MangaRepository) Subscribe(userID, providerID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(`INSERT OR REPLACE INTO manga_subscriptions (user_id, provider_id, created_at) VALUES ($1, $2, $3)`,
		userID.String(), providerID.String(), now.Format(time.RFC3339))
	return err
}

func (r *MangaRepository) Unsubscribe(userID, providerID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM manga_subscriptions WHERE user_id = $1 AND provider_id = $2`, userID.String(), providerID.String())
	return err
}

func (r *MangaRepository) GetSubscription(userID, providerID uuid.UUID) (*MangaSubscription, error) {
	row := r.db.QueryRow(`SELECT user_id, provider_id, created_at FROM manga_subscriptions WHERE user_id = $1 AND provider_id = $2`, userID.String(), providerID.String())
	var s MangaSubscription
	var createdAt string
	if err := row.Scan(&s.UserID, &s.ProviderID, &createdAt); err != nil {
		return nil, err
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &s, nil
}

func (r *MangaRepository) GetUserSubscriptions(userID uuid.UUID) ([]*MangaSubscription, error) {
	rows, err := r.db.Query(`SELECT user_id, provider_id, created_at FROM manga_subscriptions WHERE user_id = $1 ORDER BY created_at DESC`, userID.String())
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *MangaRepository) IsSubscribed(userID, providerID uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM manga_subscriptions WHERE user_id = $1 AND provider_id = $2`, userID.String(), providerID.String()).Scan(&count)
	return count > 0, err
}

// History methods
func (r *MangaRepository) UpsertHistory(userID, providerID uuid.UUID, mangaID, chapterID string, pageIndex int, title, coverURL string) error {
	now := time.Now().UTC()
	cover := sql.NullString{String: coverURL, Valid: coverURL != ""}
	_, err := r.db.Exec(`INSERT INTO manga_history (user_id, provider_id, manga_id, chapter_id, page_index, title, cover_url, last_read_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
	err := r.db.QueryRow(`SELECT COUNT(*) FROM manga_history WHERE user_id = $1`, userID.String()).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.Query(`SELECT user_id, provider_id, manga_id, chapter_id, page_index, title, cover_url, last_read_at FROM manga_history WHERE user_id = $1 ORDER BY last_read_at DESC LIMIT $2 OFFSET $3`, userID.String(), pageSize, offset)
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
	_, err := r.db.Exec(`DELETE FROM manga_history WHERE user_id = $1 AND provider_id = $2 AND manga_id = $3`, userID.String(), providerID.String(), mangaID)
	return err
}

// Settings methods
func (r *MangaRepository) GetUserSettings(userID uuid.UUID) (*MangaUserSettings, error) {
	row := r.db.QueryRow(`SELECT user_id, schema_version, data, updated_at FROM manga_user_settings WHERE user_id = $1`, userID.String())
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
		VALUES ($1, $2, $3, $4)
		ON CONFLICT(user_id) DO UPDATE SET
			schema_version = COALESCE($5, schema_version),
			data = COALESCE($6, data),
			updated_at = $7`,
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return providers, nil
}
