package persistent

import (
	"database/sql"
	"fmt"
	"time"

	"mugiko/internal/modules/media/webnovel/dto"

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

type NovelRepository struct {
	db *sql.DB
}

func NewNovelRepository(db *sql.DB) *NovelRepository {
	return &NovelRepository{db: db}
}

func (r *NovelRepository) InitializeSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS novels (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			source_id TEXT NOT NULL,
			url TEXT NOT NULL,
			title TEXT NOT NULL,
			author TEXT,
			synopsis TEXT,
			added_by TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (added_by) REFERENCES user_auths(user_id) ON DELETE CASCADE,
			UNIQUE(source, source_id)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_novels_added_by ON novels(added_by)`,
		`CREATE INDEX IF NOT EXISTS ix_novels_source_source_id ON novels(source, source_id)`,

		`CREATE TABLE IF NOT EXISTS chapters (
			id TEXT PRIMARY KEY,
			novel_id TEXT NOT NULL,
			idx INTEGER NOT NULL,
			title TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (novel_id) REFERENCES novels(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS ix_chapters_novel_id ON chapters(novel_id)`,

		`CREATE TABLE IF NOT EXISTS paragraphs (
			id TEXT PRIMARY KEY,
			chapter_id TEXT NOT NULL,
			idx INTEGER NOT NULL,
			text TEXT NOT NULL,
			words TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (chapter_id) REFERENCES chapters(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS ix_paragraphs_chapter_id ON paragraphs(chapter_id)`,
	}

	for _, q := range queries {
		if _, err := r.db.Exec(q); err != nil {
			return fmt.Errorf("schema init failed: %w", err)
		}
	}
	return nil
}

func (r *NovelRepository) GetBySourceID(source dto.SourceEnum, sourceID string) (*Novel, error) {
	row := r.db.QueryRow(`SELECT id, source, source_id, url, title, author, synopsis, added_by, created_at FROM novels WHERE source = ? AND source_id = ?`, source, sourceID)
	var n Novel
	var createdAt string
	if err := row.Scan(&n.ID, &n.Source, &n.SourceID, &n.URL, &n.Title, &n.Author, &n.Synopsis, &n.AddedBy, &createdAt); err != nil {
		return nil, err
	}
	n.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &n, nil
}

func (r *NovelRepository) GetByID(novelID uuid.UUID) (*Novel, error) {
	row := r.db.QueryRow(`SELECT id, source, source_id, url, title, author, synopsis, added_by, created_at FROM novels WHERE id = ?`, novelID.String())
	var n Novel
	var createdAt string
	if err := row.Scan(&n.ID, &n.Source, &n.SourceID, &n.URL, &n.Title, &n.Author, &n.Synopsis, &n.AddedBy, &createdAt); err != nil {
		return nil, err
	}
	n.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &n, nil
}

func (r *NovelRepository) Create(n *Novel) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(`INSERT INTO novels (id, source, source_id, url, title, author, synopsis, added_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID.String(), n.Source, n.SourceID, n.URL, n.Title, n.Author, n.Synopsis, n.AddedBy.String(), now.Format(time.RFC3339))
	if err != nil {
		return err
	}
	n.CreatedAt = now
	return nil
}

func (r *NovelRepository) ListByUser(userID uuid.UUID, page, pageSize int) ([]*Novel, int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM novels WHERE added_by = ?`, userID.String()).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(`SELECT id, source, source_id, url, title, author, synopsis, added_by, created_at FROM novels WHERE added_by = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, userID.String(), pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var novels []*Novel
	for rows.Next() {
		var n Novel
		var createdAt string
		if err := rows.Scan(&n.ID, &n.Source, &n.SourceID, &n.URL, &n.Title, &n.Author, &n.Synopsis, &n.AddedBy, &createdAt); err != nil {
			return nil, 0, err
		}
		n.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		novels = append(novels, &n)
	}
	return novels, total, nil
}

func (r *NovelRepository) ListAll(page, pageSize int) ([]*Novel, int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM novels`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(`SELECT id, source, source_id, url, title, author, synopsis, added_by, created_at FROM novels ORDER BY created_at DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var novels []*Novel
	for rows.Next() {
		var n Novel
		var createdAt string
		if err := rows.Scan(&n.ID, &n.Source, &n.SourceID, &n.URL, &n.Title, &n.Author, &n.Synopsis, &n.AddedBy, &createdAt); err != nil {
			return nil, 0, err
		}
		n.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		novels = append(novels, &n)
	}
	return novels, total, nil
}

func (r *NovelRepository) ToDTO(n *Novel) *dto.NovelDTO {
	return &dto.NovelDTO{
		ID:        n.ID,
		Source:    n.Source,
		SourceID:  n.SourceID,
		URL:       n.URL,
		Title:     n.Title,
		Author:    n.Author.String,
		Synopsis:  n.Synopsis.String,
		AddedBy:   n.AddedBy,
		CreatedAt: n.CreatedAt,
	}
}

type ChapterRepository struct {
	db *sql.DB
}

func NewChapterRepository(db *sql.DB) *ChapterRepository {
	return &ChapterRepository{db: db}
}

func (r *ChapterRepository) GetByID(chapterID uuid.UUID) (*Chapter, error) {
	row := r.db.QueryRow(`SELECT id, novel_id, idx, title, created_at FROM chapters WHERE id = ?`, chapterID.String())
	var c Chapter
	var createdAt string
	if err := row.Scan(&c.ID, &c.NovelID, &c.Idx, &c.Title, &createdAt); err != nil {
		return nil, err
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &c, nil
}

func (r *ChapterRepository) GetByNovelAndIdx(novelID uuid.UUID, idx int) (*Chapter, error) {
	row := r.db.QueryRow(`SELECT id, novel_id, idx, title, created_at FROM chapters WHERE novel_id = ? AND idx = ?`, novelID.String(), idx)
	var c Chapter
	var createdAt string
	if err := row.Scan(&c.ID, &c.NovelID, &c.Idx, &c.Title, &createdAt); err != nil {
		return nil, err
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &c, nil
}

func (r *ChapterRepository) Create(c *Chapter) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(`INSERT INTO chapters (id, novel_id, idx, title, created_at) VALUES (?, ?, ?, ?, ?)`,
		c.ID.String(), c.NovelID.String(), c.Idx, c.Title, now.Format(time.RFC3339))
	if err != nil {
		return err
	}
	c.CreatedAt = now
	return nil
}

func (r *ChapterRepository) ListByNovel(novelID uuid.UUID) ([]*Chapter, error) {
	rows, err := r.db.Query(`SELECT id, novel_id, idx, title, created_at FROM chapters WHERE novel_id = ? ORDER BY idx`, novelID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chapters []*Chapter
	for rows.Next() {
		var c Chapter
		var createdAt string
		if err := rows.Scan(&c.ID, &c.NovelID, &c.Idx, &c.Title, &createdAt); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		chapters = append(chapters, &c)
	}
	return chapters, nil
}

func (r *ChapterRepository) ToDTO(c *Chapter) *dto.ChapterDTO {
	return &dto.ChapterDTO{
		ID:        c.ID,
		NovelID:   c.NovelID,
		Idx:       c.Idx,
		Title:     c.Title.String,
		CreatedAt: c.CreatedAt,
	}
}

type ParagraphRepository struct {
	db *sql.DB
}

func NewParagraphRepository(db *sql.DB) *ParagraphRepository {
	return &ParagraphRepository{db: db}
}

func (r *ParagraphRepository) CreateBatch(chapterID uuid.UUID, paragraphs []*Paragraph) error {
	now := time.Now().UTC()
	for _, p := range paragraphs {
		_, err := r.db.Exec(`INSERT INTO paragraphs (id, chapter_id, idx, text, words, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			p.ID.String(), chapterID.String(), p.Idx, p.Text, p.Words, now.Format(time.RFC3339))
		if err != nil {
			return err
		}
		p.CreatedAt = now
	}
	return nil
}

func (r *ParagraphRepository) GetByChapter(chapterID uuid.UUID) ([]*Paragraph, error) {
	rows, err := r.db.Query(`SELECT id, chapter_id, idx, text, words, created_at FROM paragraphs WHERE chapter_id = ? ORDER BY idx`, chapterID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paragraphs []*Paragraph
	for rows.Next() {
		var p Paragraph
		var createdAt string
		if err := rows.Scan(&p.ID, &p.ChapterID, &p.Idx, &p.Text, &p.Words, &createdAt); err != nil {
			return nil, err
		}
		p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		paragraphs = append(paragraphs, &p)
	}
	return paragraphs, nil
}

func (r *ParagraphRepository) GetByChapterPaginated(chapterID uuid.UUID, limit, offset int) ([]*Paragraph, error) {
	rows, err := r.db.Query(`SELECT id, chapter_id, idx, text, words, created_at FROM paragraphs WHERE chapter_id = ? ORDER BY idx LIMIT ? OFFSET ?`, chapterID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paragraphs []*Paragraph
	for rows.Next() {
		var p Paragraph
		var createdAt string
		if err := rows.Scan(&p.ID, &p.ChapterID, &p.Idx, &p.Text, &p.Words, &createdAt); err != nil {
			return nil, err
		}
		p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		paragraphs = append(paragraphs, &p)
	}
	return paragraphs, nil
}

func (r *ParagraphRepository) CountByChapter(chapterID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM paragraphs WHERE chapter_id = ?`, chapterID.String()).Scan(&count)
	return count, err
}

func (r *ParagraphRepository) ToDTO(p *Paragraph) *dto.ParagraphDTO {
	var words []string
	// parse JSON - simplified
	return &dto.ParagraphDTO{
		ID:        p.ID,
		ChapterID: p.ChapterID,
		Idx:       p.Idx,
		Text:      p.Text,
		Words:     words,
	}
}