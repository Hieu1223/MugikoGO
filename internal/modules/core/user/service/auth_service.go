package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"mugiko/internal/modules/core/user/dto"
	"mugiko/internal/modules/core/user/exception"
	"mugiko/internal/modules/core/user/persistent"

	"github.com/google/uuid"
)

type AuthService struct {
	repo *persistent.UserRepository
}

func NewAuthService() *AuthService {
	return &AuthService{repo: persistent.NewUserRepository()}
}

func (s *AuthService) VerifyFirebaseToken(idToken string) (*FirebaseUserInfo, error) {
	// In a real implementation, this would verify the Firebase ID token
	// For now, we'll return a mock response for testing
	// TODO: Implement actual Firebase token verification
	return nil, errors.New("Firebase token verification not implemented")
}

func (s *AuthService) VerifyAdminToken(token string) (*dto.AuthData, error) {
	fmt.Printf("VerifyAdminToken: token=%s\n", token)
	auth, err := s.repo.GetUserAuthByAuthID(dto.AuthTypeAdmin, token)
	fmt.Printf("VerifyAdminToken: auth=%+v, err=%v\n", auth, err)
	if err != nil {
		return nil, err
	}
	if auth == nil || !auth.IsActive || auth.Role != dto.RoleAdmin {
		return nil, nil
	}
	return auth, nil
}

func (s *AuthService) VerifyAnonymousToken(token string) (*dto.AuthData, error) {
	auth, err := s.repo.GetUserAuthByAuthID(dto.AuthTypeAnonymous, token)
	if err != nil {
		return nil, err
	}
	if auth == nil || !auth.IsActive {
		return nil, nil
	}
	return auth, nil
}

func (s *AuthService) CreateAnonymousUser() (*dto.AuthData, string, error) {
	token := generateRandomToken(32)
	userID := uuid.New()

	err := s.repo.CreateUserAuth(userID, dto.RoleAnonymous, dto.AuthTypeAnonymous, token)
	if err != nil {
		return nil, "", err
	}

	auth := &dto.AuthData{
		UserID:   userID,
		Role:     dto.RoleAnonymous,
		JoinedAt: time.Now(),
		AuthID:   token,
		IsActive: true,
	}

	return auth, "anonymous-" + token, nil
}

func (s *AuthService) RegisterUser(request dto.RegisterRequest, firebaseUser *FirebaseUserInfo) (*dto.RegisterResponse, error) {
	// Check if Firebase UID already exists
	existing, err := s.repo.GetUserAuthByAuthID(dto.AuthTypeFirebase, firebaseUser.UID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New(exception.MsgUserExists)
	}

	userID := uuid.New()

	// Create user_auth
	err = s.repo.CreateUserAuth(userID, dto.RoleUser, dto.AuthTypeFirebase, firebaseUser.UID)
	if err != nil {
		return nil, err
	}

	// Create user_profile
	err = s.repo.CreateUserProfile(userID, request.Name, request.Bio, request.Email)
	if err != nil {
		return nil, err
	}

	token := "user-" + firebaseUser.UID

	return &dto.RegisterResponse{
		UserID: userID,
		Role:   string(dto.RoleUser),
		Token:  token,
	}, nil
}

func (s *AuthService) GetUserAuth(userID uuid.UUID) (*dto.AuthData, error) {
	return s.repo.GetUserAuthByID(userID)
}

type FirebaseUserInfo struct {
	UID   string
	Email string
}

func generateRandomToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based token
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(bytes)
}
