package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"mugiko/internal/modules/media/subtitle/dto"
	"mugiko/internal/modules/media/subtitle/persistent"
)

// SubtitleService handles subtitle operations
type SubtitleService struct {
	repo *persistent.SubtitleRepository
}

func NewSubtitleService(repo *persistent.SubtitleRepository) *SubtitleService {
	return &SubtitleService{repo: repo}
}

// Provider methods
func (s *SubtitleService) ListProviders() ([]*dto.SubtitleProviderResponse, error) {
	providers, err := s.repo.GetActiveProviders()
	if err != nil {
		return nil, err
	}

	result := make([]*dto.SubtitleProviderResponse, len(providers))
	for i, p := range providers {
		services, _ := s.repo.GetProviderServices(p.ID)
		serviceResponses := make([]dto.SubtitleProviderServiceResponse, len(services))
		for j, svc := range services {
			var options any
			if svc.Options != nil {
				json.Unmarshal([]byte(*svc.Options), &options)
			}
			serviceResponses[j] = dto.SubtitleProviderServiceResponse{
				ID:         svc.ID,
				ProviderID: svc.ProviderID,
				ServiceKey: svc.ServiceKey,
				Name:       svc.Name,
				Options:    options,
				IsActive:   svc.IsActive,
				CreatedAt:  svc.CreatedAt,
				UpdatedAt:  svc.UpdatedAt,
			}
		}

		var baseURL string
		if p.BaseURL != nil {
			baseURL = *p.BaseURL
		}
		var lastSyncedAt *time.Time
		if p.LastSyncedAt != nil {
			lastSyncedAt = p.LastSyncedAt
		}

		result[i] = &dto.SubtitleProviderResponse{
			ID:           p.ID,
			Name:         p.Name,
			Kind:         dto.ProviderKind(p.Kind),
			BaseURL:      baseURL,
			IsActive:     p.IsActive,
			LastSyncedAt: lastSyncedAt,
			CreatedAt:    p.CreatedAt,
			UpdatedAt:    p.UpdatedAt,
			Services:     serviceResponses,
		}
	}

	return result, nil
}

func (s *SubtitleService) CreateProvider(data dto.SubtitleProviderCreate) (*dto.SubtitleProviderResponse, error) {
	var baseURL string
	if data.BaseURL != "" {
		baseURL = data.BaseURL
	}

	provider, err := s.repo.CreateProvider(data.Name, persistent.ProviderKind(data.Kind), &baseURL)
	if err != nil {
		return nil, err
	}

	return &dto.SubtitleProviderResponse{
		ID:        provider.ID,
		Name:      provider.Name,
		Kind:      dto.ProviderKind(provider.Kind),
		BaseURL:   baseURL,
		IsActive:  provider.IsActive,
		CreatedAt: provider.CreatedAt,
		UpdatedAt: provider.UpdatedAt,
		Services:  []dto.SubtitleProviderServiceResponse{},
	}, nil
}

// Subtitle methods
func (s *SubtitleService) CreateSubtitle(userID uuid.UUID, data dto.SubtitleCreate) (*dto.SubtitleResponse, error) {
	// Extract video ID from YouTube URL
	videoID := extractVideoID(data.VideoURL)
	if videoID == "" {
		return nil, fmt.Errorf("invalid video URL")
	}

	// Verify provider exists
	_, err := s.repo.GetProvider(data.ProviderID)
	if err != nil {
		return nil, persistent.ErrProviderNotFound
	}

	// Verify service exists
	service, err := s.repo.GetProviderService(data.ProviderID, data.ServiceKey)
	if err != nil {
		return nil, persistent.ErrServiceNotFound
	}

	// Create subtitle record
	subtitle, err := s.repo.CreateSubtitle(videoID, data.VideoURL, data.ProviderID, service.ID, userID)
	if err != nil {
		return nil, err
	}

	// Return response (segments empty initially)
	return &dto.SubtitleResponse{
		ID:         subtitle.ID,
		VideoID:    subtitle.VideoID,
		VideoURL:   subtitle.VideoURL,
		ProviderID: subtitle.ProviderID,
		ServiceID:  subtitle.ServiceID,
		CreatedBy:  subtitle.CreatedBy,
		Status:     subtitle.Status,
		CreatedAt:  subtitle.CreatedAt,
		UpdatedAt:  subtitle.UpdatedAt,
	}, nil
}

func (s *SubtitleService) GetSubtitle(subtitleID uuid.UUID, userID *uuid.UUID) (*dto.SubtitleResponse, error) {
	subtitle, err := s.repo.GetSubtitle(subtitleID)
	if err != nil {
		return nil, err
	}

	// Record access if user provided
	if userID != nil {
		s.repo.UpsertAccess(*userID, subtitleID)
	}

	return &dto.SubtitleResponse{
		ID:         subtitle.ID,
		VideoID:    subtitle.VideoID,
		VideoURL:   subtitle.VideoURL,
		ProviderID: subtitle.ProviderID,
		ServiceID:  subtitle.ServiceID,
		CreatedBy:  subtitle.CreatedBy,
		Status:     subtitle.Status,
		CreatedAt:  subtitle.CreatedAt,
		UpdatedAt:  subtitle.UpdatedAt,
	}, nil
}

func (s *SubtitleService) GetSubtitleSegments(subtitleID uuid.UUID, limit, offset int) (*dto.SubtitleSegmentsResponse, error) {
	// Get total count
	total, err := s.repo.GetSegmentCount(subtitleID)
	if err != nil {
		return nil, err
	}

	segments, err := s.repo.GetSegments(subtitleID, limit, offset)
	if err != nil {
		return nil, err
	}

	segmentResponses := make([]dto.SubtitleSegmentResponse, len(segments))
	for i, seg := range segments {
		var words any
		if seg.Words != nil {
			json.Unmarshal([]byte(*seg.Words), &words)
		}
		segmentResponses[i] = dto.SubtitleSegmentResponse{
			ID:         seg.ID,
			SubtitleID: seg.SubtitleID,
			Idx:        seg.Idx,
			StartMs:    seg.StartMs,
			EndMs:      seg.EndMs,
			Words:      words,
		}
	}

	return &dto.SubtitleSegmentsResponse{
		Items:  segmentResponses,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *SubtitleService) SearchSubtitles(videoURL string, limit, offset int) ([]*dto.SubtitleListResponse, error) {
	videoID := extractVideoID(videoURL)
	if videoID == "" {
		return nil, fmt.Errorf("invalid video URL")
	}

	subtitles, err := s.repo.GetSubtitlesByVideoID(videoID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.SubtitleListResponse, len(subtitles))
	for i, sub := range subtitles {
		// Get segment count
		segs, _ := s.repo.GetSegments(sub.ID, 1000, 0)
		result[i] = &dto.SubtitleListResponse{
			ID:           sub.ID,
			VideoID:      sub.VideoID,
			VideoURL:     sub.VideoURL,
			ProviderID:   sub.ProviderID,
			ServiceID:    sub.ServiceID,
			CreatedBy:    sub.CreatedBy,
			CreatedAt:    sub.CreatedAt,
			SegmentCount: len(segs),
		}
	}

	return result, nil
}

// User settings
func (s *SubtitleService) GetUserSettings(userID uuid.UUID) (*dto.UserSettingsResponse, error) {
	settings, err := s.repo.GetUserSettings(userID)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		return &dto.UserSettingsResponse{
			UserID:        userID,
			SchemaVersion: 1,
			Data:          map[string]any{},
			UpdatedAt:     time.Now(),
		}, nil
	}

	var data any
	json.Unmarshal([]byte(settings.Data), &data)

	return &dto.UserSettingsResponse{
		UserID:        settings.UserID,
		SchemaVersion: settings.SchemaVersion,
		Data:          data,
		UpdatedAt:     settings.UpdatedAt,
	}, nil
}

func (s *SubtitleService) UpdateUserSettings(userID uuid.UUID, data dto.UserSettingsUpdate) (*dto.UserSettingsResponse, error) {
	settings, err := s.repo.UpsertUserSettings(userID, data.SchemaVersion, data.Data)
	if err != nil {
		return nil, err
	}

	var dataMap any
	json.Unmarshal([]byte(settings.Data), &dataMap)

	return &dto.UserSettingsResponse{
		UserID:        settings.UserID,
		SchemaVersion: settings.SchemaVersion,
		Data:          dataMap,
		UpdatedAt:     settings.UpdatedAt,
	}, nil
}

// Subtitle user state
func (s *SubtitleService) GetSubtitleUserState(userID, subtitleID uuid.UUID) (*dto.SubtitleUserStateResponse, error) {
	state, err := s.repo.GetUserState(userID, subtitleID)
	if err != nil {
		return nil, err
	}

	if state == nil {
		return &dto.SubtitleUserStateResponse{
			UserID:        userID,
			SubtitleID:    subtitleID,
			SchemaVersion: 1,
			Data:          map[string]any{},
			UpdatedAt:     time.Now(),
		}, nil
	}

	var data any
	json.Unmarshal([]byte(state.Data), &data)

	return &dto.SubtitleUserStateResponse{
		UserID:        state.UserID,
		SubtitleID:    state.SubtitleID,
		SchemaVersion: state.SchemaVersion,
		Data:          data,
		UpdatedAt:     state.UpdatedAt,
	}, nil
}

func (s *SubtitleService) UpdateSubtitleUserState(userID, subtitleID uuid.UUID, data dto.SubtitleUserStateUpdate) (*dto.SubtitleUserStateResponse, error) {
	state, err := s.repo.UpsertUserState(userID, subtitleID, data.SchemaVersion, data.Data)
	if err != nil {
		return nil, err
	}

	var dataMap any
	json.Unmarshal([]byte(state.Data), &dataMap)

	return &dto.SubtitleUserStateResponse{
		UserID:        state.UserID,
		SubtitleID:    state.SubtitleID,
		SchemaVersion: state.SchemaVersion,
		Data:          dataMap,
		UpdatedAt:     state.UpdatedAt,
	}, nil
}

// Subtitle History (visit history)
func (s *SubtitleService) GetUserHistory(userID uuid.UUID, limit, offset int) (*dto.SubtitleHistoryListResponse, error) {
	items, total, err := s.repo.GetUserHistory(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.SubtitleHistoryResponse, len(items))
	for i, item := range items {
		responses[i] = dto.SubtitleHistoryResponse{
			SubtitleID:     item.SubtitleID,
			VideoID:        item.VideoID,
			VideoURL:       item.VideoURL,
			LastAccessedAt: item.LastAccessedAt,
			SegmentCount:   item.SegmentCount,
		}
	}

	return &dto.SubtitleHistoryListResponse{
		Items:  responses,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// Helper function to extract YouTube video ID
func extractVideoID(url string) string {
	// Simple extraction for YouTube URLs
	// Supports: youtube.com/watch?v=..., youtu.be/..., youtube.com/embed/..., youtube.com/shorts/...
	if len(url) == 0 {
		return ""
	}

	// youtube.com/watch?v=VIDEO_ID
	if idx := len("youtube.com/watch?v="); len(url) > idx {
		for i := 0; i <= len(url)-idx; i++ {
			if url[i:i+idx] == "youtube.com/watch?v=" {
				start := i + idx
				end := start
				for end < len(url) && url[end] != '&' && url[end] != '#' && url[end] != '?' {
					end++
				}
				if end > start && end-start <= 20 {
					return url[start:end]
				}
			}
		}
	}

	// youtu.be/VIDEO_ID
	if idx := len("youtu.be/"); len(url) > idx {
		for i := 0; i <= len(url)-idx; i++ {
			if url[i:i+idx] == "youtu.be/" {
				start := i + idx
				end := start
				for end < len(url) && url[end] != '&' && url[end] != '#' && url[end] != '?' && url[end] != '/' {
					end++
				}
				if end > start && end-start <= 20 {
					return url[start:end]
				}
			}
		}
	}

	// youtube.com/embed/VIDEO_ID
	if idx := len("youtube.com/embed/"); len(url) > idx {
		for i := 0; i <= len(url)-idx; i++ {
			if url[i:i+idx] == "youtube.com/embed/" {
				start := i + idx
				end := start
				for end < len(url) && url[end] != '&' && url[end] != '#' && url[end] != '?' && url[end] != '/' {
					end++
				}
				if end > start && end-start <= 20 {
					return url[start:end]
				}
			}
		}
	}

	// youtube.com/shorts/VIDEO_ID
	if idx := len("youtube.com/shorts/"); len(url) > idx {
		for i := 0; i <= len(url)-idx; i++ {
			if url[i:i+idx] == "youtube.com/shorts/" {
				start := i + idx
				end := start
				for end < len(url) && url[end] != '&' && url[end] != '#' && url[end] != '?' && url[end] != '/' {
					end++
				}
				if end > start && end-start <= 20 {
					return url[start:end]
				}
			}
		}
	}

	return ""
}
