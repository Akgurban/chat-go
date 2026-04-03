package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"chat-go/internal/models"
	"chat-go/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Username == "" || req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Username, email and password are required")
		return
	}

	if len(req.Password) < 6 {
		respondWithError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	user, token, err := h.authService.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			respondWithError(w, http.StatusConflict, "User already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"user":  user.ToResponse(),
		"token": token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, token, err := h.authService.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to login")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"user":  user.ToResponse(),
		"token": token,
	})
}
