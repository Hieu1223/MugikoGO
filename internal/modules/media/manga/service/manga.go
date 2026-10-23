package service

import (
	"database/sql"
	"net/http"

	commonexception "mugiko/internal/common/exception"
	"mugiko/internal/modules/core/user/middleware"
	"mugiko/internal/modules/media/manga/dto"
	"mugiko/internal/modules/media/manga/exception"
	"mugiko/internal/modules/media/manga/persistent"

	"github.com/google/uuid"
)

type MangaService struct {
	repo *persistent.MangaRepository
}

func NewMangaService(repo *persistent.MangaRepository) *MangaService {
	return &MangaService{repo: repo}
}

// Provider methods
func (s *MangaService) GetUserProviders(r *http.Request, userID uuid.UUID) ([]*persistent.MangaProvider, error) {
	providers, err := s.repo.GetUserProviders(userID)
	if err != nil {
		return nil, err
	}
	return providers, nil
}

func (s *MangaService) CreateProvider(r *http.Request, req dto.CreateProviderRequest) (*dto.MangaProviderDTO, error) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	// Check if provider already exists
	existing, err := s.repo.GetProviderByOwnerAndURL(userID, req.BaseURL)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if existing != nil {
		return nil, ErrProviderAlreadyExists
	}

	vis := req.Visibility
	if vis == "" {
		vis = dto.ProviderVisibilityPrivate
	}

	provider, err := s.repo.CreateProvider(userID, req.Name, req.BaseURL, vis, false)
	if err != nil {
		return nil, err
	}

	return s.ToProviderDTO(provider), nil
}

func (s *MangaService) GetProvider(providerID uuid.UUID) (*dto.MangaProviderDTO, error) {
	provider, err := s.repo.GetProvider(providerID)
	if err != nil {
		return nil, err
	}
	return s.ToProviderDTO(provider), nil
}

func (s *MangaService) ListPublicProviders(page, pageSize int) ([]*dto.MangaProviderDTO, int, error) {
	providers, total, err := s.repo.GetPublicProviders(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return s.ToProviderDTOs(providers), total, nil
}

func (s *MangaService) GetSubscribableProviders(r *http.Request, userID uuid.UUID) ([]*dto.MangaProviderDTO, error) {
	providers, err := s.repo.GetSubscribableProviders(userID)
	if err != nil {
		return nil, err
	}
	return s.ToProviderDTOs(providers), nil
}

func (s *MangaService) UpdateProvider(r *http.Request, providerID uuid.UUID, req dto.UpdateProviderRequest) (*dto.MangaProviderDTO, error) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	provider, err := s.repo.GetProvider(providerID)
	if err != nil {
		return nil, err
	}
	if provider.OwnerID != userID && user.Role != "admin" {
		return nil, ErrForbidden
	}

	if req.Name != nil {
		provider.Name = *req.Name
	}
	if req.BaseURL != nil {
		provider.BaseURL = *req.BaseURL
	}
	if req.Visibility != nil {
		provider.Visibility = *req.Visibility
	}

	if err := s.repo.UpdateProvider(provider); err != nil {
		return nil, err
	}
	return s.ToProviderDTO(provider), nil
}

func (s *MangaService) DeleteProvider(r *http.Request, providerID uuid.UUID) error {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return ErrUnauthorized
	}

	provider, err := s.repo.GetProvider(providerID)
	if err != nil {
		return err
	}
	if provider.OwnerID != userID && user.Role != "admin" {
		return ErrForbidden
	}

	return s.repo.DeleteProvider(providerID)
}

// Subscription methods
func (s *MangaService) Subscribe(r *http.Request, providerID uuid.UUID) error {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return ErrUnauthorized
	}

	// Check provider exists and is subscribable
	provider, err := s.repo.GetProvider(providerID)
	if err != nil {
		return err
	}
	if provider.Visibility != dto.ProviderVisibilityPublic && provider.OwnerID != userID {
		return ErrForbidden
	}

	return s.repo.Subscribe(userID, providerID)
}

func (s *MangaService) Unsubscribe(r *http.Request, providerID uuid.UUID) error {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return ErrUnauthorized
	}
	return s.repo.Unsubscribe(userID, providerID)
}

func (s *MangaService) ListSubscriptions(r *http.Request) ([]*dto.MangaSubscriptionDTO, error) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}
	subs, err := s.repo.GetUserSubscriptions(userID)
	if err != nil {
		return nil, err
	}
	var result []*dto.MangaSubscriptionDTO
	for _, sub := range subs {
		provider, _ := s.repo.GetProvider(sub.ProviderID)
		result = append(result, &dto.MangaSubscriptionDTO{
			UserID:     sub.UserID,
			ProviderID: sub.ProviderID,
			Provider:   s.ToProviderDTO(provider),
			CreatedAt:  sub.CreatedAt,
		})
	}
	return result, nil
}

func (s *MangaService) IsSubscribed(r *http.Request, providerID uuid.UUID) (bool, error) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return false, ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return false, ErrUnauthorized
	}
	return s.repo.IsSubscribed(userID, providerID)
}

// History methods
func (s *MangaService) UpsertHistory(r *http.Request, req dto.UpsertHistoryRequest) error {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return ErrUnauthorized
	}
	return s.repo.UpsertHistory(userID, req.ProviderID, req.MangaID, req.ChapterID, req.PageIndex, req.Title, req.CoverURL)
}

func (s *MangaService) GetUserHistory(r *http.Request, page, pageSize int) ([]*dto.MangaHistoryDTO, int, error) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return nil, 0, ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return nil, 0, ErrUnauthorized
	}
	history, total, err := s.repo.GetUserHistory(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	var result []*dto.MangaHistoryDTO
	for _, h := range history {
		provider, _ := s.repo.GetProvider(h.ProviderID)
		result = append(result, &dto.MangaHistoryDTO{
			UserID:     h.UserID,
			ProviderID: h.ProviderID,
			MangaID:    h.MangaID,
			ChapterID:  h.ChapterID,
			PageIndex:  h.PageIndex,
			Title:      h.Title,
			CoverURL:   h.CoverURL.String,
			LastReadAt: h.LastReadAt,
			Provider:   s.ToProviderDTO(provider),
		})
	}
	return result, total, nil
}

func (s *MangaService) DeleteHistory(r *http.Request, providerID uuid.UUID, mangaID string) error {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return ErrUnauthorized
	}
	return s.repo.DeleteHistory(userID, providerID, mangaID)
}

// Settings methods
func (s *MangaService) GetUserSettings(r *http.Request) (*dto.MangaUserSettingsDTO, error) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}
	settings, err := s.repo.GetUserSettings(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &dto.MangaUserSettingsDTO{UserID: userID, SchemaVersion: 1, Data: map[string]interface{}{}}, nil
		}
		return nil, err
	}
	var data map[string]interface{}
	if len(settings.Data) > 0 {
		// parse JSON - simplified
	}
	return &dto.MangaUserSettingsDTO{
		UserID:        settings.UserID,
		SchemaVersion: settings.SchemaVersion,
		Data:          data,
		UpdatedAt:     settings.UpdatedAt,
	}, nil
}

func (s *MangaService) UpdateUserSettings(r *http.Request, req dto.UpdateSettingsRequest) (*dto.MangaUserSettingsDTO, error) {
	user, ok := middleware.GetUserInfo(r.Context())
	if !ok || user == (middleware.UserInfo{}) {
		return nil, ErrUnauthorized
	}
	userID, err := uuid.Parse(user.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}
	settings, err := s.repo.UpsertUserSettings(userID, req.SchemaVersion, req.Data)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	if len(settings.Data) > 0 {
		// parse JSON - simplified
	}
	return &dto.MangaUserSettingsDTO{
		UserID:        settings.UserID,
		SchemaVersion: settings.SchemaVersion,
		Data:          data,
		UpdatedAt:     settings.UpdatedAt,
	}, nil
}

// Helper methods
func (s *MangaService) ToProviderDTO(p *persistent.MangaProvider) *dto.MangaProviderDTO {
	return &dto.MangaProviderDTO{
		ID:         p.ID,
		OwnerID:    p.OwnerID,
		Name:       p.Name,
		BaseURL:    p.BaseURL,
		Visibility: p.Visibility,
		IsBuiltin:  p.IsBuiltin,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

func (s *MangaService) ToProviderDTOs(providers []*persistent.MangaProvider) []*dto.MangaProviderDTO {
	var result []*dto.MangaProviderDTO
	for _, p := range providers {
		result = append(result, s.ToProviderDTO(p))
	}
	return result
}

// Errors
var (
	ErrUnauthorized          = &ServiceError{Code: "UNAUTHORIZED", Status: http.StatusUnauthorized, Message: commonexception.MsgUnauthorized}
	ErrForbidden             = &ServiceError{Code: "FORBIDDEN", Status: http.StatusForbidden, Message: exception.MsgAccessDenied}
	ErrProviderAlreadyExists = &ServiceError{Code: "PROVIDER_EXISTS", Status: http.StatusConflict, Message: exception.MsgProviderAlreadyExists}
	ErrMangaNotFound         = &ServiceError{Code: "MANGA_NOT_FOUND", Status: http.StatusNotFound, Message: exception.MsgMangaNotFound}
	ErrNotImplemented        = &ServiceError{Code: "NOT_IMPLEMENTED", Status: http.StatusNotImplemented, Message: exception.MsgNotImplemented}
)

type ServiceError struct {
	Code    string
	Status  int
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}
