package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"study-tracker-backend/internal/auth"
	"study-tracker-backend/internal/repository"
)

type AuthHandler struct {
	users *repository.UserRepository
	auth  *auth.Service
}

func NewAuthHandler(users *repository.UserRepository, authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		users: users,
		auth:  authService,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, passwordHash, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to authenticate")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.auth.CreateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create token")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: token})
}
