package persistent

import (
	"database/sql"
	"mugiko/internal/modules/media/manga/dto"
	"time"

	"github.com/google/uuid"
)

type MangaProvider struct {
	ID         uuid.UUID
	OwnerID    uuid.UUID
	Name       string
	BaseURL    string
	Visibility dto.ProviderVisibility
	IsBuiltin  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type MangaSubscription struct {
	UserID     uuid.UUID
	ProviderID uuid.UUID
	CreatedAt  time.Time
}

type MangaHistory struct {
	UserID     uuid.UUID
	ProviderID uuid.UUID
	MangaID    string
	ChapterID  string
	PageIndex  int
	Title      string
	CoverURL   sql.NullString
	LastReadAt time.Time
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
