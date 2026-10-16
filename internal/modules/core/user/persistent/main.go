package persistent

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"mugiko/internal/common/persistent"
	"mugiko/internal/modules/core/user/dto"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{db: persistent.GetDB()}
}

func (r *UserRepository) CreateUserAuth(userID uuid.UUID, role dto.Role, authType dto.AuthType, authID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	query := `INSERT INTO users (user_id, role, auth_type, auth_id, is_active, joined_at) 
	          VALUES ($1, $2, $3, $4, 1, $5)`
	_, err := r.db.Exec(query, userID.String(), role, authType, authID, now)
	return err
}

func (r *UserRepository) CreateUserProfile(userID uuid.UUID, name, bio, email string) error {
	query := `INSERT INTO user_profiles (user_id, name, bio, email) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, userID.String(), name, bio, email)
	return err
}

func (r *UserRepository) GetUserAuthByAuthID(authType dto.AuthType, authID string) (*dto.AuthData, error) {
	query := `SELECT user_id, role, joined_at, auth_id, is_active FROM users WHERE auth_type = $1 AND auth_id = $2`
	row := r.db.QueryRow(query, authType, authID)

	var userIDStr, roleStr, authIDStr, joinedAtStr string
	var isActiveInt int // Schema uses INTEGER for is_active

	err := row.Scan(&userIDStr, &roleStr, &joinedAtStr, &authIDStr, &isActiveInt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	userID, _ := uuid.Parse(userIDStr)
	joinedAt, err := time.Parse(time.RFC3339, joinedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse joined_at: %w", err)
	}

	return &dto.AuthData{
		UserID:   userID,
		Role:     dto.Role(roleStr),
		JoinedAt: joinedAt,
		AuthID:   authIDStr,
		IsActive: isActiveInt == 1,
	}, nil
}

func (r *UserRepository) GetUserAuthByID(userID uuid.UUID) (*dto.AuthData, error) {
	query := `SELECT user_id, role, joined_at, auth_id, is_active FROM users WHERE user_id = $1`
	row := r.db.QueryRow(query, userID.String())

	var userIDStr, roleStr, authIDStr, joinedAtStr string
	var isActiveInt int

	err := row.Scan(&userIDStr, &roleStr, &joinedAtStr, &authIDStr, &isActiveInt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	joinedAt, err := time.Parse(time.RFC3339, joinedAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse joined_at: %w", err)
	}

	return &dto.AuthData{
		UserID:   userID,
		Role:     dto.Role(roleStr),
		JoinedAt: joinedAt,
		AuthID:   authIDStr,
		IsActive: isActiveInt == 1,
	}, nil
}

func (r *UserRepository) GetUserProfileByID(userID uuid.UUID) (*dto.UserProfileResponse, error) {
	query := `SELECT name, bio, email FROM user_profiles WHERE user_id = $1`
	row := r.db.QueryRow(query, userID.String())

	var name, bio, email sql.NullString
	err := row.Scan(&name, &bio, &email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &dto.UserProfileResponse{
		UserID: userID,
		Name:   name.String,
		Bio:    bio.String,
		Email:  email.String,
	}, nil
}

func (r *UserRepository) UpdateUserProfile(userID uuid.UUID, name, bio, email *string) error {
	setClauses := []string{}
	args := []any{}
	paramIdx := 1

	if name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", paramIdx))
		args = append(args, *name)
		paramIdx++
	}
	if bio != nil {
		setClauses = append(setClauses, fmt.Sprintf("bio = $%d", paramIdx))
		args = append(args, *bio)
		paramIdx++
	}
	if email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", paramIdx))
		args = append(args, *email)
		paramIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	// Build query: UPDATE user_profiles SET col=$1, col2=$2 ... WHERE user_id=$N
	query := fmt.Sprintf(`UPDATE user_profiles SET %s WHERE user_id = $%d`,
		strings.Join(setClauses, ", "), paramIdx)

	args = append(args, userID.String())

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *UserRepository) ListUsers(limit, offset int) ([]*dto.AuthData, error) {
	query := `SELECT user_id, role, joined_at, auth_id, is_active FROM users ORDER BY joined_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*dto.AuthData
	for rows.Next() {
		var userIDStr, roleStr, authIDStr, joinedAtStr string
		var isActiveInt int

		err := rows.Scan(&userIDStr, &roleStr, &joinedAtStr, &authIDStr, &isActiveInt)
		if err != nil {
			return nil, err
		}

		userID, _ := uuid.Parse(userIDStr)
		joinedAt, err := time.Parse(time.RFC3339, joinedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse joined_at: %w", err)
		}

		users = append(users, &dto.AuthData{
			UserID:   userID,
			Role:     dto.Role(roleStr),
			JoinedAt: joinedAt,
			AuthID:   authIDStr,
			IsActive: isActiveInt == 1,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) GetUserWithProfile(userID uuid.UUID) (*dto.AuthData, *dto.UserProfileResponse, error) {
	auth, err := r.GetUserAuthByID(userID)
	if err != nil || auth == nil {
		return nil, nil, err
	}
	profile, err := r.GetUserProfileByID(userID)
	return auth, profile, err
}

func (r *UserRepository) UpdateUserRole(userID uuid.UUID, role dto.Role) error {
	query := `UPDATE users SET role = $1 WHERE user_id = $2`
	_, err := r.db.Exec(query, role, userID.String())
	return err
}

func (r *UserRepository) DeactivateUser(userID uuid.UUID) error {
	query := `UPDATE users SET is_active = 0 WHERE user_id = $1`
	_, err := r.db.Exec(query, userID.String())
	return err
}

func (r *UserRepository) AdminExists() (bool, error) {
	query := `SELECT 1 FROM users WHERE role = 'admin' LIMIT 1`
	row := r.db.QueryRow(query)
	var exists int
	err := row.Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *UserRepository) CountUsers() (int, error) {
	query := `SELECT COUNT(*) FROM users`
	row := r.db.QueryRow(query)
	var count int
	err := row.Scan(&count)
	return count, err
}
