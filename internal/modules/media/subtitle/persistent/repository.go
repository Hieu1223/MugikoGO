package persistent

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type SubtitleRepository struct {
	db *sql.DB
}

func NewSubtitleRepository(db *sql.DB) *SubtitleRepository {
	return &SubtitleRepository{db: db}
}

func (r *SubtitleRepository) InitializeSchema() error {
	schema := `
	-- Subtitle providers
	CREATE TABLE IF NOT EXISTS subtitle_providers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		kind TEXT NOT NULL,
		base_url TEXT,
		is_active BOOLEAN DEFAULT TRUE NOT NULL,
		last_synced_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	-- Subtitle provider services
	CREATE TABLE IF NOT EXISTS subtitle_provider_services (
		id TEXT PRIMARY KEY,
		provider_id TEXT NOT NULL REFERENCES subtitle_providers(id) ON DELETE CASCADE,
		service_key TEXT NOT NULL,
		name TEXT NOT NULL,
		options TEXT,
		is_active BOOLEAN DEFAULT TRUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		UNIQUE(provider_id, service_key)
	);

	-- Subtitles
	CREATE TABLE IF NOT EXISTS subtitles (
		id TEXT PRIMARY KEY,
		video_id TEXT NOT NULL,
		video_url TEXT NOT NULL,
		provider_id TEXT NOT NULL REFERENCES subtitle_providers(id),
		service_id TEXT NOT NULL REFERENCES subtitle_provider_services(id),
		created_by TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS ix_subtitles_video_id ON subtitles(video_id);

	-- Subtitle segments
	CREATE TABLE IF NOT EXISTS subtitle_segments (
		id TEXT PRIMARY KEY,
		subtitle_id TEXT NOT NULL REFERENCES subtitles(id) ON DELETE CASCADE,
		idx INTEGER NOT NULL,
		start_ms INTEGER NOT NULL,
		end_ms INTEGER NOT NULL,
		words TEXT,
		UNIQUE(subtitle_id, idx)
	);

	-- Subtitle access (visit history)
	CREATE TABLE IF NOT EXISTS subtitle_access (
		user_id TEXT NOT NULL,
		subtitle_id TEXT NOT NULL REFERENCES subtitles(id) ON DELETE CASCADE,
		last_accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		PRIMARY KEY (user_id, subtitle_id)
	);

	-- User settings
	CREATE TABLE IF NOT EXISTS user_settings (
		user_id TEXT PRIMARY KEY,
		schema_version INTEGER DEFAULT 1 NOT NULL,
		data TEXT NOT NULL DEFAULT '{}',
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	-- Subtitle user state
	CREATE TABLE IF NOT EXISTS subtitle_user_state (
		user_id TEXT NOT NULL,
		subtitle_id TEXT NOT NULL REFERENCES subtitles(id) ON DELETE CASCADE,
		schema_version INTEGER DEFAULT 1 NOT NULL,
		data TEXT NOT NULL DEFAULT '{}',
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		PRIMARY KEY (user_id, subtitle_id)
	);

	-- Trigger for updated_at on subtitle_providers
	CREATE TRIGGER IF NOT EXISTS trg_subtitle_providers_updated_at
	AFTER UPDATE ON subtitle_providers
	BEGIN
		UPDATE subtitle_providers SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;

	-- Trigger for updated_at on subtitle_provider_services
	CREATE TRIGGER IF NOT EXISTS trg_subtitle_provider_services_updated_at
	AFTER UPDATE ON subtitle_provider_services
	BEGIN
		UPDATE subtitle_provider_services SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;

	-- Trigger for updated_at on user_settings
	CREATE TRIGGER IF NOT EXISTS trg_user_settings_updated_at
	AFTER UPDATE ON user_settings
	BEGIN
		UPDATE user_settings SET updated_at = CURRENT_TIMESTAMP WHERE user_id = NEW.user_id;
	END;

	-- Trigger for updated_at on subtitle_user_state
	CREATE TRIGGER IF NOT EXISTS trg_subtitle_user_state_updated_at
	AFTER UPDATE ON subtitle_user_state
	BEGIN
		UPDATE subtitle_user_state SET updated_at = CURRENT_TIMESTAMP WHERE user_id = NEW.user_id AND subtitle_id = NEW.subtitle_id;
	END;
	`
	_, err := r.db.Exec(schema)
	return err
}

// Provider methods
func (r *SubtitleRepository) CreateProvider(name string, kind ProviderKind, baseURL *string) (*SubtitleProvider, error) {
	id := uuid.New()
	now := time.Now()
	
	var baseURLStr string
	if baseURL != nil {
		baseURLStr = *baseURL
	}
	
	_, err := r.db.Exec(
		`INSERT INTO subtitle_providers (id, name, kind, base_url, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id.String(), name, string(kind), baseURLStr, true, now, now,
	)
	if err != nil {
		return nil, err
	}
	
	return &SubtitleProvider{
		ID:        id,
		Name:      name,
		Kind:      kind,
		BaseURL:   baseURL,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (r *SubtitleRepository) GetProvider(providerID uuid.UUID) (*SubtitleProvider, error) {
	row := r.db.QueryRow(
		`SELECT id, name, kind, base_url, is_active, last_synced_at, created_at, updated_at FROM subtitle_providers WHERE id = ?`,
		providerID.String(),
	)
	
	var p SubtitleProvider
	var baseURL sql.NullString
	var lastSyncedAt sql.NullTime
	var kindStr string
	
	err := row.Scan(&p.ID, &p.Name, &kindStr, &baseURL, &p.IsActive, &lastSyncedAt, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrProviderNotFound
	}
	if err != nil {
		return nil, err
	}
	
	p.Kind = ProviderKind(kindStr)
	if baseURL.Valid {
		p.BaseURL = &baseURL.String
	}
	if lastSyncedAt.Valid {
		p.LastSyncedAt = &lastSyncedAt.Time
	}
	
	// Load services
	services, err := r.GetProviderServices(providerID)
	if err != nil {
		return nil, err
	}
	// Note: we'd need to add Services field to SubtitleProvider if we want to return them
	_ = services
	
	return &p, nil
}

func (r *SubtitleRepository) GetActiveProviders() ([]*SubtitleProvider, error) {
	rows, err := r.db.Query(
		`SELECT id, name, kind, base_url, is_active, last_synced_at, created_at, updated_at FROM subtitle_providers WHERE is_active = TRUE`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var providers []*SubtitleProvider
	for rows.Next() {
		var p SubtitleProvider
		var baseURL sql.NullString
		var lastSyncedAt sql.NullTime
		var kindStr string
		
		err := rows.Scan(&p.ID, &p.Name, &kindStr, &baseURL, &p.IsActive, &lastSyncedAt, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		
		p.Kind = ProviderKind(kindStr)
		if baseURL.Valid {
			p.BaseURL = &baseURL.String
		}
		if lastSyncedAt.Valid {
			p.LastSyncedAt = &lastSyncedAt.Time
		}
		
		providers = append(providers, &p)
	}
	
	return providers, rows.Err()
}

func (r *SubtitleRepository) UpdateProviderSyncTime(providerID uuid.UUID) error {
	_, err := r.db.Exec(
		`UPDATE subtitle_providers SET last_synced_at = CURRENT_TIMESTAMP WHERE id = ?`,
		providerID.String(),
	)
	return err
}

func (r *SubtitleRepository) GetProviderServices(providerID uuid.UUID) ([]*SubtitleProviderService, error) {
	rows, err := r.db.Query(
		`SELECT id, provider_id, service_key, name, options, is_active, created_at, updated_at FROM subtitle_provider_services WHERE provider_id = ? AND is_active = TRUE`,
		providerID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var services []*SubtitleProviderService
	for rows.Next() {
		var s SubtitleProviderService
		var options sql.NullString
		
		err := rows.Scan(&s.ID, &s.ProviderID, &s.ServiceKey, &s.Name, &options, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		
		if options.Valid {
			s.Options = &options.String
		}
		
		services = append(services, &s)
	}
	
	return services, rows.Err()
}

func (r *SubtitleRepository) GetProviderService(providerID uuid.UUID, serviceKey string) (*SubtitleProviderService, error) {
	row := r.db.QueryRow(
		`SELECT id, provider_id, service_key, name, options, is_active, created_at, updated_at FROM subtitle_provider_services WHERE provider_id = ? AND service_key = ? AND is_active = TRUE`,
		providerID.String(), serviceKey,
	)
	
	var s SubtitleProviderService
	var options sql.NullString
	
	err := row.Scan(&s.ID, &s.ProviderID, &s.ServiceKey, &s.Name, &options, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrServiceNotFound
	}
	if err != nil {
		return nil, err
	}
	
	if options.Valid {
		s.Options = &options.String
	}
	
	return &s, nil
}

func (r *SubtitleRepository) UpsertServices(providerID uuid.UUID, services []map[string]any) error {
	// Deactivate all existing services
	_, err := r.db.Exec(
		`UPDATE subtitle_provider_services SET is_active = FALSE WHERE provider_id = ?`,
		providerID.String(),
	)
	if err != nil {
		return err
	}
	
	for _, svc := range services {
		id := uuid.New()
		now := time.Now()
		
		var optionsStr string
		if opts, ok := svc["options"]; ok && opts != nil {
			data, _ := json.Marshal(opts)
			optionsStr = string(data)
		}
		
		_, err := r.db.Exec(
			`INSERT INTO subtitle_provider_services (id, provider_id, service_key, name, options, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id.String(), providerID.String(), svc["service_key"].(string), svc["name"].(string), optionsStr, true, now, now,
		)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// Subtitle methods
func (r *SubtitleRepository) CreateSubtitle(videoID, videoURL string, providerID, serviceID, createdBy uuid.UUID) (*Subtitle, error) {
	id := uuid.New()
	now := time.Now()
	
	_, err := r.db.Exec(
		`INSERT INTO subtitles (id, video_id, video_url, provider_id, service_id, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id.String(), videoID, videoURL, providerID.String(), serviceID.String(), createdBy.String(), now,
	)
	if err != nil {
		return nil, err
	}
	
	return &Subtitle{
		ID:         id,
		VideoID:    videoID,
		VideoURL:   videoURL,
		ProviderID: providerID,
		ServiceID:  serviceID,
		CreatedBy:  createdBy,
		CreatedAt:  now,
	}, nil
}

func (r *SubtitleRepository) GetSubtitle(subtitleID uuid.UUID) (*Subtitle, error) {
	row := r.db.QueryRow(
		`SELECT id, video_id, video_url, provider_id, service_id, created_by, created_at FROM subtitles WHERE id = ?`,
		subtitleID.String(),
	)
	
	var s Subtitle
	err := row.Scan(&s.ID, &s.VideoID, &s.VideoURL, &s.ProviderID, &s.ServiceID, &s.CreatedBy, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrSubtitleNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return &s, nil
}

func (r *SubtitleRepository) GetSubtitlesByVideoID(videoID string, limit, offset int) ([]*Subtitle, error) {
	rows, err := r.db.Query(
		`SELECT id, video_id, video_url, provider_id, service_id, created_by, created_at FROM subtitles WHERE video_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		videoID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var subtitles []*Subtitle
	for rows.Next() {
		var s Subtitle
		err := rows.Scan(&s.ID, &s.VideoID, &s.VideoURL, &s.ProviderID, &s.ServiceID, &s.CreatedBy, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		subtitles = append(subtitles, &s)
	}
	
	return subtitles, rows.Err()
}

func (r *SubtitleRepository) AddSegments(subtitleID uuid.UUID, segments []map[string]any) error {
	for idx, seg := range segments {
		id := uuid.New()
		
		var wordsStr string
		if words, ok := seg["words"]; ok && words != nil {
			data, _ := json.Marshal(words)
			wordsStr = string(data)
		}
		
		_, err := r.db.Exec(
			`INSERT INTO subtitle_segments (id, subtitle_id, idx, start_ms, end_ms, words) VALUES (?, ?, ?, ?, ?, ?)`,
			id.String(), subtitleID.String(), idx, seg["start_ms"], seg["end_ms"], wordsStr,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *SubtitleRepository) GetSegments(subtitleID uuid.UUID, limit, offset int) ([]*SubtitleSegment, error) {
	rows, err := r.db.Query(
		`SELECT id, subtitle_id, idx, start_ms, end_ms, words FROM subtitle_segments WHERE subtitle_id = ? ORDER BY idx LIMIT ? OFFSET ?`,
		subtitleID.String(), limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []*SubtitleSegment
	for rows.Next() {
		var s SubtitleSegment
		var words sql.NullString
		
		err := rows.Scan(&s.ID, &s.SubtitleID, &s.Idx, &s.StartMs, &s.EndMs, &words)
		if err != nil {
			return nil, err
		}
		
		if words.Valid {
			s.Words = &words.String
		}
		
		segments = append(segments, &s)
	}
	
	return segments, rows.Err()
}

func (r *SubtitleRepository) GetSegmentCount(subtitleID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM subtitle_segments WHERE subtitle_id = ?`,
		subtitleID.String(),
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Subtitle access (visit history)
func (r *SubtitleRepository) UpsertAccess(userID, subtitleID uuid.UUID) error {
	now := time.Now()
	
	_, err := r.db.Exec(
		`INSERT INTO subtitle_access (user_id, subtitle_id, last_accessed_at) VALUES (?, ?, ?)
		 ON CONFLICT(user_id, subtitle_id) DO UPDATE SET last_accessed_at = ?`,
		userID.String(), subtitleID.String(), now, now,
	)
	return err
}

func (r *SubtitleRepository) GetUserHistory(userID uuid.UUID, limit, offset int) ([]*SubtitleHistoryItem, int, error) {
	// Get total count
	var total int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM subtitle_access sa
		 JOIN subtitles s ON sa.subtitle_id = s.id
		 WHERE sa.user_id = ?`,
		userID.String(),
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated history with subtitle details
	rows, err := r.db.Query(
		`SELECT sa.subtitle_id, s.video_id, s.video_url, sa.last_accessed_at,
		        (SELECT COUNT(*) FROM subtitle_segments WHERE subtitle_id = sa.subtitle_id) as segment_count
		 FROM subtitle_access sa
		 JOIN subtitles s ON sa.subtitle_id = s.id
		 WHERE sa.user_id = ?
		 ORDER BY sa.last_accessed_at DESC
		 LIMIT ? OFFSET ?`,
		userID.String(), limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var history []*SubtitleHistoryItem
	for rows.Next() {
		var item SubtitleHistoryItem
		err := rows.Scan(&item.SubtitleID, &item.VideoID, &item.VideoURL, &item.LastAccessedAt, &item.SegmentCount)
		if err != nil {
			return nil, 0, err
		}
		history = append(history, &item)
	}

	return history, total, rows.Err()
}

type SubtitleHistoryItem struct {
	SubtitleID     uuid.UUID
	VideoID        string
	VideoURL       string
	LastAccessedAt time.Time
	SegmentCount   int
}

// User settings
func (r *SubtitleRepository) GetUserSettings(userID uuid.UUID) (*UserSettings, error) {
	row := r.db.QueryRow(
		`SELECT user_id, schema_version, data, updated_at FROM user_settings WHERE user_id = ?`,
		userID.String(),
	)
	
	var s UserSettings
	err := row.Scan(&s.UserID, &s.SchemaVersion, &s.Data, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &s, nil
}

func (r *SubtitleRepository) UpsertUserSettings(userID uuid.UUID, schemaVersion int, data any) (*UserSettings, error) {
	dataStr := "{}"
	if data != nil {
		dataBytes, _ := json.Marshal(data)
		dataStr = string(dataBytes)
	}
	
	now := time.Now()
	
	_, err := r.db.Exec(
		`INSERT INTO user_settings (user_id, schema_version, data, updated_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET schema_version = ?, data = ?, updated_at = ?`,
		userID.String(), schemaVersion, dataStr, now, schemaVersion, dataStr, now,
	)
	if err != nil {
		return nil, err
	}
	
	return &UserSettings{
		UserID:        userID,
		SchemaVersion: schemaVersion,
		Data:          dataStr,
		UpdatedAt:     now,
	}, nil
}

// Subtitle user state
func (r *SubtitleRepository) GetUserState(userID, subtitleID uuid.UUID) (*SubtitleUserState, error) {
	row := r.db.QueryRow(
		`SELECT user_id, subtitle_id, schema_version, data, updated_at FROM subtitle_user_state WHERE user_id = ? AND subtitle_id = ?`,
		userID.String(), subtitleID.String(),
	)
	
	var s SubtitleUserState
	err := row.Scan(&s.UserID, &s.SubtitleID, &s.SchemaVersion, &s.Data, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	return &s, nil
}

func (r *SubtitleRepository) UpsertUserState(userID, subtitleID uuid.UUID, schemaVersion int, data any) (*SubtitleUserState, error) {
	dataStr := "{}"
	if data != nil {
		dataBytes, _ := json.Marshal(data)
		dataStr = string(dataBytes)
	}
	
	now := time.Now()
	
	_, err := r.db.Exec(
		`INSERT INTO subtitle_user_state (user_id, subtitle_id, schema_version, data, updated_at) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(user_id, subtitle_id) DO UPDATE SET schema_version = ?, data = ?, updated_at = ?`,
		userID.String(), subtitleID.String(), schemaVersion, dataStr, now, schemaVersion, dataStr, now,
	)
	if err != nil {
		return nil, err
	}
	
	return &SubtitleUserState{
		UserID:        userID,
		SubtitleID:    subtitleID,
		SchemaVersion: schemaVersion,
		Data:          dataStr,
		UpdatedAt:     now,
	}, nil
}