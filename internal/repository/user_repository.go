package repository

import (
	"database/sql"
	"errors"

	"chat-go/internal/models"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserExists = errors.New("user already exists")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		"offline",
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar_url, status, last_seen_at, created_at, updated_at
		FROM users WHERE id = $1`

	user := &models.User{}
	var lastSeenAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Status,
		&lastSeenAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if lastSeenAt.Valid {
		user.LastSeenAt = &lastSeenAt.Time
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar_url, status, last_seen_at, created_at, updated_at
		FROM users WHERE email = $1`

	user := &models.User{}
	var lastSeenAt sql.NullTime
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Status,
		&lastSeenAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if lastSeenAt.Valid {
		user.LastSeenAt = &lastSeenAt.Time
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar_url, status, last_seen_at, created_at, updated_at
		FROM users WHERE username = $1`

	user := &models.User{}
	var lastSeenAt sql.NullTime
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Status,
		&lastSeenAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if lastSeenAt.Valid {
		user.LastSeenAt = &lastSeenAt.Time
	}
	return user, nil
}

func (r *UserRepository) UpdateStatus(userID int, status string) error {
	if status == "offline" {
		query := `UPDATE users SET status = $1, last_seen_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
		_, err := r.db.Exec(query, status, userID)
		return err
	}
	query := `UPDATE users SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.Exec(query, status, userID)
	return err
}

func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar_url, status, last_seen_at, created_at, updated_at
		FROM users ORDER BY username`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var lastSeenAt sql.NullTime
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.AvatarURL,
			&user.Status,
			&lastSeenAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if lastSeenAt.Valid {
			user.LastSeenAt = &lastSeenAt.Time
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) UsernameExists(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	err := r.db.QueryRow(query, username).Scan(&exists)
	return exists, err
}

// SearchUsers searches for users by email or username (partial match)
func (r *UserRepository) SearchUsers(query string, limit int) ([]models.User, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	searchQuery := `
		SELECT id, username, email, password_hash, avatar_url, status, last_seen_at, created_at, updated_at
		FROM users 
		WHERE LOWER(username) LIKE LOWER($1) OR LOWER(email) LIKE LOWER($1)
		ORDER BY 
			CASE 
				WHEN LOWER(username) = LOWER($2) THEN 1
				WHEN LOWER(email) = LOWER($2) THEN 2
				WHEN LOWER(username) LIKE LOWER($2 || '%') THEN 3
				WHEN LOWER(email) LIKE LOWER($2 || '%') THEN 4
				ELSE 5
			END,
			username
		LIMIT $3`

	searchPattern := "%" + query + "%"
	rows, err := r.db.Query(searchQuery, searchPattern, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var lastSeenAt sql.NullTime
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.AvatarURL,
			&user.Status,
			&lastSeenAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if lastSeenAt.Valid {
			user.LastSeenAt = &lastSeenAt.Time
		}
		users = append(users, user)
	}
	return users, nil
}

// GetByEmailOrUsername finds a user by exact email or username match
func (r *UserRepository) GetByEmailOrUsername(identifier string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar_url, status, last_seen_at, created_at, updated_at
		FROM users 
		WHERE LOWER(email) = LOWER($1) OR LOWER(username) = LOWER($1)`

	user := &models.User{}
	var lastSeenAt sql.NullTime
	err := r.db.QueryRow(query, identifier).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Status,
		&lastSeenAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if lastSeenAt.Valid {
		user.LastSeenAt = &lastSeenAt.Time
	}
	return user, nil
}
