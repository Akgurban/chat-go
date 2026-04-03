package handler

import (
	"net/http"
	"strconv"
	"strings"

	"chat-go/internal/repository"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := h.userRepo.GetAllUsers()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	// Convert to response format (hide password hashes)
	var userResponses []map[string]interface{}
	for _, u := range users {
		userResponses = append(userResponses, map[string]interface{}{
			"id":         u.ID,
			"username":   u.Username,
			"email":      u.Email,
			"avatar_url": u.AvatarURL,
			"status":     u.Status,
			"created_at": u.CreatedAt,
		})
	}

	if userResponses == nil {
		userResponses = []map[string]interface{}{}
	}

	respondWithJSON(w, http.StatusOK, userResponses)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	userID, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	respondWithJSON(w, http.StatusOK, user.ToResponse())
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id").(int)

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	respondWithJSON(w, http.StatusOK, user.ToResponse())
}
