package service

import (
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	commonexception "mugiko/internal/common/exception"
	"mugiko/internal/modules/core/user/middleware"
	"mugiko/internal/modules/media/webnovel/dto"
	"mugiko/internal/modules/media/webnovel/exception"
	"mugiko/internal/modules/media/webnovel/persistent"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New(commonexception.MsgUnauthorized)
	ErrNotFound     = errors.New(exception.MsgNotFound)
	ErrInvalidURL   = errors.New(exception.MsgInvalidURL)
	ErrNovelExists  = errors.New(exception.MsgNovelExists)
	ErrSyncNotImpl  = errors.New(exception.MsgSyncNotImplemented)
)

type NovelService struct {
	novelRepo     *persistent.NovelRepository
	chapterRepo   *persistent.ChapterRepository
	paragraphRepo *persistent.ParagraphRepository
}

func NewNovelService(novelRepo *persistent.NovelRepository, chapterRepo *persistent.ChapterRepository, paragraphRepo *persistent.ParagraphRepository) *NovelService {
	return &NovelService{
		novelRepo:     novelRepo,
		chapterRepo:   chapterRepo,
		paragraphRepo: paragraphRepo,
	}
}

func (s *NovelService) Preview(r *http.Request, req dto.PreviewRequest) (*dto.PreviewResponse, error) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return nil, ErrUnauthorized
	}

	// Validate and extract source from URL
	source, sourceID, err := s.parseSyosetsuURL(req.URL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	// Check if novel already exists
	existing, err := s.novelRepo.GetBySourceID(source, sourceID)
	if err == nil && existing != nil {
		return &dto.PreviewResponse{
			URL:      existing.URL,
			Title:    existing.Title,
			Author:   existing.Author.String,
			Synopsis: existing.Synopsis.String,
			Source:   existing.Source,
		}, nil
	}

	// Fetch from external source (simplified - would call actual API)
	preview, err := s.fetchPreview(source, sourceID)
	if err != nil {
		return nil, err
	}
	return preview, nil
}

func (s *NovelService) AddNovel(r *http.Request, req dto.AddNovelRequest) (*dto.AddNovelResponse, error) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	source, sourceID, err := s.parseSyosetsuURL(req.URL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	// Check if already exists
	existing, err := s.novelRepo.GetBySourceID(source, sourceID)
	if err == nil && existing != nil {
		return &dto.AddNovelResponse{
			Novel: s.novelRepo.ToDTO(existing),
			JobID: "",
		}, nil
	}

	// Fetch preview data
	preview, err := s.fetchPreview(source, sourceID)
	if err != nil {
		return nil, err
	}

	// Create novel
	novel := &persistent.Novel{
		ID:        uuid.New(),
		Source:    source,
		SourceID:  sourceID,
		URL:       req.URL,
		Title:     preview.Title,
		Author:    sqlNullString(preview.Author),
		Synopsis:  sqlNullString(preview.Synopsis),
		AddedBy:   userID,
	}

	if err := s.novelRepo.Create(novel); err != nil {
		return nil, err
	}

	// Sync not implemented per user request
	// Would trigger background sync job here

	return &dto.AddNovelResponse{
		Novel: s.novelRepo.ToDTO(novel),
		JobID: "", // No sync job for now
	}, nil
}

func (s *NovelService) ListNovelsByUser(r *http.Request, page, pageSize int) (*dto.PaginatedNovelsResponse, error) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	novels, total, err := s.novelRepo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	var result []*dto.NovelDTO
	for _, n := range novels {
		result = append(result, s.novelRepo.ToDTO(n))
	}

	return &dto.PaginatedNovelsResponse{
		Novels:    result,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

func (s *NovelService) ListAllNovels(page, pageSize int) (*dto.PaginatedNovelsResponse, error) {
	novels, total, err := s.novelRepo.ListAll(page, pageSize)
	if err != nil {
		return nil, err
	}

	var result []*dto.NovelDTO
	for _, n := range novels {
		result = append(result, s.novelRepo.ToDTO(n))
	}

	return &dto.PaginatedNovelsResponse{
		Novels:    result,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

func (s *NovelService) GetNovel(novelID uuid.UUID) (*dto.NovelWithChaptersResponse, error) {
	novel, err := s.novelRepo.GetByID(novelID)
	if err != nil {
		return nil, err
	}

	chapters, err := s.chapterRepo.ListByNovel(novelID)
	if err != nil {
		return nil, err
	}

	var chapterDTOs []*dto.ChapterDTO
	for _, c := range chapters {
		chapterDTOs = append(chapterDTOs, s.chapterRepo.ToDTO(c))
	}

	return &dto.NovelWithChaptersResponse{
		NovelDTO:  s.novelRepo.ToDTO(novel),
		Chapters:  chapterDTOs,
	}, nil
}

func (s *NovelService) GetChapter(chapterID uuid.UUID) (*dto.ChapterResponse, error) {
	chapter, err := s.chapterRepo.GetByID(chapterID)
	if err != nil {
		return nil, err
	}

	paragraphs, err := s.paragraphRepo.GetByChapter(chapterID)
	if err != nil {
		return nil, err
	}

	var paraDTOs []*dto.ParagraphDTO
	for _, p := range paragraphs {
		paraDTOs = append(paraDTOs, s.paragraphRepo.ToDTO(p))
	}

	return &dto.ChapterResponse{
		ChapterDTO:  s.chapterRepo.ToDTO(chapter),
		Paragraphs:  paraDTOs,
	}, nil
}

func (s *NovelService) GetChapterPaginated(chapterID uuid.UUID, limit, offset int) (*dto.ChapterResponse, error) {
	chapter, err := s.chapterRepo.GetByID(chapterID)
	if err != nil {
		return nil, err
	}

	paragraphs, err := s.paragraphRepo.GetByChapterPaginated(chapterID, limit, offset)
	if err != nil {
		return nil, err
	}

	var paraDTOs []*dto.ParagraphDTO
	for _, p := range paragraphs {
		paraDTOs = append(paraDTOs, s.paragraphRepo.ToDTO(p))
	}

	return &dto.ChapterResponse{
		ChapterDTO:  s.chapterRepo.ToDTO(chapter),
		Paragraphs:  paraDTOs,
	}, nil
}

func (s *NovelService) parseSyosetsuURL(rawURL string) (dto.SourceEnum, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}

	if !strings.Contains(u.Host, "syosetu.com") && !strings.Contains(u.Host, "ncode.syosetu.com") {
		return "", "", errors.New(exception.MsgUnsupportedSource)
	}

	// Extract ncode from path
	// e.g., /n12345ab/ or /n12345ab
	re := regexp.MustCompile(`/n([a-z0-9]+)/?`)
	matches := re.FindStringSubmatch(u.Path)
	if len(matches) < 2 {
		return "", "", errors.New(exception.MsgCouldNotExtractNcode)
	}

	return dto.SourceSyosetsu, "n" + matches[1], nil
}

func (s *NovelService) fetchPreview(source dto.SourceEnum, sourceID string) (*dto.PreviewResponse, error) {
	// Simplified - in reality would call Syosetsu API
	// For now return placeholder
	return &dto.PreviewResponse{
		URL:      "https://ncode.syosetu.com/" + sourceID + "/",
		Title:    "Sample Novel " + sourceID,
		Author:   "Sample Author",
		Synopsis: "Sample synopsis for " + sourceID,
		Source:   source,
	}, nil
}

func sqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func (s *NovelService) Search(query string, limit int) (*dto.SearchResponse, error) {
	// Simplified - would search external API
	// For now return empty results
	return &dto.SearchResponse{
		Items: []*dto.NovelDTO{},
		Total: 0,
	}, nil
}

func (s *NovelService) Sync(novelID uuid.UUID, force bool) (string, error) {
	return "", ErrSyncNotImpl
}