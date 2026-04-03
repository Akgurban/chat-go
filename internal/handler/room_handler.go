package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"chat-go/internal/models"
	"chat-go/internal/repository"
)

type RoomHandler struct {
	roomRepo *repository.RoomRepository
}

func NewRoomHandler(roomRepo *repository.RoomRepository) *RoomHandler {
	return &RoomHandler{roomRepo: roomRepo}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	var req models.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Room name is required")
		return
	}

	room := &models.Room{
		Name:        req.Name,
		Description: &req.Description,
		IsPrivate:   req.IsPrivate,
		CreatedBy:   &userID,
	}

	if err := h.roomRepo.Create(room); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create room")
		return
	}

	// Add creator as admin member
	if err := h.roomRepo.AddMember(room.ID, userID, "admin"); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to add member to room")
		return
	}

	respondWithJSON(w, http.StatusCreated, room)
}

func (h *RoomHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rooms, err := h.roomRepo.GetAllPublic()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get rooms")
		return
	}

	if rooms == nil {
		rooms = []models.Room{}
	}

	respondWithJSON(w, http.StatusOK, rooms)
}

func (h *RoomHandler) GetMyRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	rooms, err := h.roomRepo.GetUserRooms(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get rooms")
		return
	}

	if rooms == nil {
		rooms = []models.Room{}
	}

	respondWithJSON(w, http.StatusOK, rooms)
}

func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract room ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	roomID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	room, err := h.roomRepo.GetByID(roomID)
	if err != nil {
		if err == repository.ErrRoomNotFound {
			respondWithError(w, http.StatusNotFound, "Room not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get room")
		return
	}

	members, err := h.roomRepo.GetMembers(roomID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get room members")
		return
	}

	// Convert to user responses
	memberResponses := make([]models.UserResponse, len(members))
	for i, m := range members {
		memberResponses[i] = m.ToResponse()
	}

	response := models.RoomWithMembers{
		Room:    *room,
		Members: memberResponses,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract room ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	roomID, err := strconv.Atoi(parts[len(parts)-2])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	// Check if room exists
	room, err := h.roomRepo.GetByID(roomID)
	if err != nil {
		if err == repository.ErrRoomNotFound {
			respondWithError(w, http.StatusNotFound, "Room not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get room")
		return
	}

	// Check if room is private
	if room.IsPrivate {
		respondWithError(w, http.StatusForbidden, "Cannot join private room")
		return
	}

	if err := h.roomRepo.AddMember(roomID, userID, "member"); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to join room")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Joined room successfully"})
}

func (h *RoomHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Extract room ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	roomID, err := strconv.Atoi(parts[len(parts)-2])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	if err := h.roomRepo.RemoveMember(roomID, userID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to leave room")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Left room successfully"})
}
