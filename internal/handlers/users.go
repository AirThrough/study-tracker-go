package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"study-tracker-backend/internal/auth"
	"study-tracker-backend/internal/repository"
)

type UserHandler struct {
	users *repository.UserRepository
	auth  *auth.Service
}

func NewUserHandler(users *repository.UserRepository, authService *auth.Service) *UserHandler {
	return &UserHandler{
		users: users,
		auth:  authService,
	}
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type updateUserRequest struct {
	Email *string `json:"email"`
	Name  *string `json:"name"`
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "email, password, and name are required")
		return
	}

	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user, err := h.users.Create(r.Context(), repository.CreateUserInput{
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		Name:         req.Name,
	})
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			writeError(w, http.StatusConflict, "email already in use")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, err := h.auth.CreateToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create token")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":  user,
		"token": token,
	})
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.users.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	currentUserID, ok := auth.UserIDFromContext(r.Context())
	if !ok || currentUserID != id {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == nil && req.Name == nil {
		writeError(w, http.StatusBadRequest, "at least one field is required")
		return
	}

	user, err := h.users.Update(r.Context(), id, repository.UpdateUserInput{
		Email: req.Email,
		Name:  req.Name,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, repository.ErrEmailTaken) {
			writeError(w, http.StatusConflict, "email already in use")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	currentUserID, ok := auth.UserIDFromContext(r.Context())
	if !ok || currentUserID != id {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.users.SoftDelete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	writeJSON(w, http.StatusOK, user)
}
