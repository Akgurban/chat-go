package repository

import (
	"database/sql"
	"errors"

	"chat-go/internal/models"
)

var ErrRoomNotFound = errors.New("room not found")

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(room *models.Room) error {
	query := `
		INSERT INTO rooms (name, description, is_private, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(
		query,
		room.Name,
		room.Description,
		room.IsPrivate,
		room.CreatedBy,
	).Scan(&room.ID, &room.CreatedAt, &room.UpdatedAt)
}

func (r *RoomRepository) GetByID(id int) (*models.Room, error) {
	query := `
		SELECT id, name, description, is_private, created_by, created_at, updated_at
		FROM rooms WHERE id = $1`

	room := &models.Room{}
	err := r.db.QueryRow(query, id).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.IsPrivate,
		&room.CreatedBy,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *RoomRepository) GetAllPublic() ([]models.Room, error) {
	query := `
		SELECT id, name, description, is_private, created_by, created_at, updated_at
		FROM rooms WHERE is_private = false ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Description,
			&room.IsPrivate,
			&room.CreatedBy,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *RoomRepository) GetUserRooms(userID int) ([]models.Room, error) {
	query := `
		SELECT r.id, r.name, r.description, r.is_private, r.created_by, r.created_at, r.updated_at
		FROM rooms r
		INNER JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = $1
		ORDER BY r.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Description,
			&room.IsPrivate,
			&room.CreatedBy,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *RoomRepository) AddMember(roomID, userID int, role string) error {
	query := `
		INSERT INTO room_members (room_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (room_id, user_id) DO NOTHING`

	_, err := r.db.Exec(query, roomID, userID, role)
	return err
}

func (r *RoomRepository) RemoveMember(roomID, userID int) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	_, err := r.db.Exec(query, roomID, userID)
	return err
}

func (r *RoomRepository) IsMember(roomID, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)`
	var exists bool
	err := r.db.QueryRow(query, roomID, userID).Scan(&exists)
	return exists, err
}

func (r *RoomRepository) GetMembers(roomID int) ([]models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.avatar_url, u.status, u.created_at, u.updated_at
		FROM users u
		INNER JOIN room_members rm ON u.id = rm.user_id
		WHERE rm.room_id = $1`

	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.AvatarURL,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *RoomRepository) Delete(roomID int) error {
	query := `DELETE FROM rooms WHERE id = $1`
	_, err := r.db.Exec(query, roomID)
	return err
}

func (r *RoomRepository) GetMemberIDs(roomID int) ([]int, error) {
	query := `SELECT user_id FROM room_members WHERE room_id = $1`

	rows, err := r.db.Query(query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}
