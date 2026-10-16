package persistent

import (
	"database/sql"
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
	query := `INSERT INTO user_auths (user_id, role, auth_type, auth_id, is_active) VALUES (?, ?, ?, ?, 1)`
	_, err := r.db.Exec(query, userID.String(), role, authType, authID)
	return err
}

func (r *UserRepository) CreateUserProfile(userID uuid.UUID, name, bio, email string) error {
	query := `INSERT INTO user_profiles (user_id, name, bio, email) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, userID.String(), name, bio, email)
	return err
}

func (r *UserRepository) GetUserAuthByAuthID(authType dto.AuthType, authID string) (*dto.AuthData, error) {
	query := `SELECT user_id, role, joined_at, auth_id, is_active FROM user_auths WHERE auth_type = ? AND auth_id = ?`
	row := r.db.QueryRow(query, authType, authID)

	var userIDStr, roleStr, authIDStr string
	var joinedAt time.Time
	var isActive bool

	err := row.Scan(&userIDStr, &roleStr, &joinedAt, &authIDStr, &isActive)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	userID, _ := uuid.Parse(userIDStr)
	return &dto.AuthData{
		UserID:   userID,
		Role:     dto.Role(roleStr),
		JoinedAt: joinedAt,
		AuthID:   authIDStr,
		IsActive: isActive,
	}, nil
}

func (r *UserRepository) GetUserAuthByID(userID uuid.UUID) (*dto.AuthData, error) {
	query := `SELECT user_id, role, joined_at, auth_id, is_active FROM user_auths WHERE user_id = ?`
	row := r.db.QueryRow(query, userID.String())

	var userIDStr, roleStr, authIDStr string
	var joinedAt time.Time
	var isActive bool

	err := row.Scan(&userIDStr, &roleStr, &joinedAt, &authIDStr, &isActive)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &dto.AuthData{
		UserID:   userID,
		Role:     dto.Role(roleStr),
		JoinedAt: joinedAt,
		AuthID:   authIDStr,
		IsActive: isActive,
	}, nil
}

func (r *UserRepository) GetUserProfileByID(userID uuid.UUID) (*dto.UserProfileResponse, error) {
	query := `SELECT name, bio, email FROM user_profiles WHERE user_id = ?`
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
	query := `UPDATE user_profiles SET `
	args := []any{}
	setClauses := []string{}

	if name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *name)
	}
	if bio != nil {
		setClauses = append(setClauses, "bio = ?")
		args = append(args, *bio)
	}
	if email != nil {
		setClauses = append(setClauses, "email = ?")
		args = append(args, *email)
	}

	if len(setClauses) == 0 {
		return nil
	}

	query += setClauses[0]
	for i := 1; i < len(setClauses); i++ {
		query += ", " + setClauses[i]
	}
	query += " WHERE user_id = ?"
	args = append(args, userID.String())

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *UserRepository) ListUsers(limit, offset int) ([]*dto.AuthData, error) {
	query := `SELECT user_id, role, joined_at, auth_id, is_active FROM user_auths ORDER BY joined_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*dto.AuthData
	for rows.Next() {
		var userIDStr, roleStr, authIDStr string
		var joinedAt time.Time
		var isActive bool

		err := rows.Scan(&userIDStr, &roleStr, &joinedAt, &authIDStr, &isActive)
		if err != nil {
			return nil, err
		}

		userID, _ := uuid.Parse(userIDStr)
		users = append(users, &dto.AuthData{
			UserID:   userID,
			Role:     dto.Role(roleStr),
			JoinedAt: joinedAt,
			AuthID:   authIDStr,
			IsActive: isActive,
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
	query := `UPDATE user_auths SET role = ? WHERE user_id = ?`
	_, err := r.db.Exec(query, role, userID.String())
	return err
}

func (r *UserRepository) DeactivateUser(userID uuid.UUID) error {
	query := `UPDATE user_auths SET is_active = 0 WHERE user_id = ?`
	_, err := r.db.Exec(query, userID.String())
	return err
}

func (r *UserRepository) AdminExists() (bool, error) {
	query := `SELECT 1 FROM user_auths WHERE role = 'admin' LIMIT 1`
	row := r.db.QueryRow(query)
	var exists int
	err := row.Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *UserRepository) CountUsers() (int, error) {
	query := `SELECT COUNT(*) FROM user_auths`
	row := r.db.QueryRow(query)
	var count int
	err := row.Scan(&count)
	return count, err
}
