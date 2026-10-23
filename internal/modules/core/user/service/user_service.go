package service

import (
	"errors"
	"mugiko/internal/modules/core/user/dto"
	"mugiko/internal/modules/core/user/exception"
	"mugiko/internal/modules/core/user/persistent"

	"github.com/google/uuid"
)

type UserService struct {
	repo *persistent.UserRepository
}

func NewUserService() *UserService {
	return &UserService{repo: persistent.NewUserRepository()}
}

func (s *UserService) GetMyProfile(userID uuid.UUID) (*dto.UserProfileResponse, error) {
	auth, err := s.repo.GetUserAuthByID(userID)
	if err != nil || auth == nil || !auth.IsActive {
		return nil, err
	}

	profile, err := s.repo.GetUserProfileByID(userID)
	if err != nil || profile == nil {
		return nil, err
	}

	profile.UserID = auth.UserID
	profile.Role = string(auth.Role)
	profile.JoinedAt = auth.JoinedAt

	return profile, nil
}

func (s *UserService) UpdateMyProfile(userID uuid.UUID, name, bio, email *string) (*dto.UserProfileResponse, error) {
	auth, err := s.repo.GetUserAuthByID(userID)
	if err != nil || auth == nil || !auth.IsActive {
		return nil, err
	}

	profile, err := s.repo.GetUserProfileByID(userID)
	if err != nil || profile == nil {
		return nil, err
	}

	err = s.repo.UpdateUserProfile(userID, name, bio, email)
	if err != nil {
		return nil, err
	}

	// Get updated profile
	updatedProfile, err := s.repo.GetUserProfileByID(userID)
	if err != nil || updatedProfile == nil {
		return nil, err
	}

	updatedProfile.UserID = auth.UserID
	updatedProfile.Role = string(auth.Role)
	updatedProfile.JoinedAt = auth.JoinedAt

	return updatedProfile, nil
}

func (s *UserService) ListUsers(limit, offset int) ([]*dto.AuthData, error) {
	return s.repo.ListUsers(limit, offset)
}

func (s *UserService) GetUserDetail(userID uuid.UUID) (*dto.AuthData, *dto.UserProfileResponse, error) {
	return s.repo.GetUserWithProfile(userID)
}

func (s *UserService) UpdateUserRole(userID uuid.UUID, role dto.Role) error {
	// Validate role
	if role != dto.RoleAnonymous && role != dto.RoleUser && role != dto.RoleAdmin {
		return errors.New(exception.MsgInvalidRole)
	}
	return s.repo.UpdateUserRole(userID, role)
}

func (s *UserService) DeactivateUser(userID uuid.UUID) error {
	return s.repo.DeactivateUser(userID)
}
