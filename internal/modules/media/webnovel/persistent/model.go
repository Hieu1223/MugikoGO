package persistent

import (
	"database/sql"
	"mugiko/internal/modules/media/webnovel/dto"
	"time"

	"github.com/google/uuid"
)

type Novel struct {
	ID        uuid.UUID
	Source    dto.SourceEnum
	SourceID  string
	URL       string
	Title     string
	Author    sql.NullString
	Synopsis  sql.NullString
	AddedBy   uuid.UUID
	CreatedAt time.Time
}

type Chapter struct {
	ID        uuid.UUID
	NovelID   uuid.UUID
	Idx       int
	Title     sql.NullString
	CreatedAt time.Time
}

type Paragraph struct {
	ID        uuid.UUID
	ChapterID uuid.UUID
	Idx       int
	Text      string
	Words     []byte // JSON
	CreatedAt time.Time
}
