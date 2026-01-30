package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"study-tracker-backend/internal/apperrors"
	"study-tracker-backend/internal/auth"
	"study-tracker-backend/internal/models"
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
	Email    string  `json:"email"`
	Password string  `json:"password"`
	Name     string  `json:"name"`
	Role     *string `json:"role"`
}

type updateUserRequest struct {
	Email *string `json:"email"`
	Name  *string `json:"name"`
	Role  *string `json:"role"`
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r.Context()) {
		writeAppError(w, apperrors.CodeForbidden, http.StatusForbidden, errors.New("admin access required"))
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, apperrors.CodeInvalidRequestBody, http.StatusBadRequest, err)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		writeAppError(w, apperrors.CodeValidationRequired, http.StatusBadRequest, errors.New("email, password, and name are required"))
		return
	}

	if len(req.Password) < 8 {
		writeAppError(w, apperrors.CodeValidationPasswordShort, http.StatusBadRequest, errors.New("password must be at least 8 characters"))
		return
	}

	role := models.RoleUser
	if req.Role != nil {
		parsedRole := models.Role(*req.Role)
		if !parsedRole.IsValid() {
			writeAppError(w, apperrors.CodeValidationInvalidRole, http.StatusBadRequest, errors.New("role must be ADMIN or USER"))
			return
		}
		role = parsedRole
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeAppError(w, apperrors.CodePasswordHashFailed, http.StatusInternalServerError, err)
		return
	}

	user, err := h.users.Create(r.Context(), repository.CreateUserInput{
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		Name:         req.Name,
		Role:         role,
	})
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			writeAppError(w, apperrors.CodeUserEmailTaken, http.StatusConflict, err)
			return
		}
		writeAppError(w, apperrors.CodeUserCreateFailed, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r.Context()) {
		writeAppError(w, apperrors.CodeForbidden, http.StatusForbidden, errors.New("admin access required"))
		return
	}

	users, err := h.users.List(r.Context())
	if err != nil {
		writeAppError(w, apperrors.CodeUserListFailed, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !canReadUser(r, id) {
		writeAppError(w, apperrors.CodeForbidden, http.StatusForbidden, errors.New("access denied"))
		return
	}

	user, err := h.users.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeAppError(w, apperrors.CodeUserNotFound, http.StatusNotFound, err)
			return
		}
		writeAppError(w, apperrors.CodeUserGetFailed, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r.Context()) {
		writeAppError(w, apperrors.CodeForbidden, http.StatusForbidden, errors.New("admin access required"))
		return
	}

	id := chi.URLParam(r, "id")

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, apperrors.CodeInvalidRequestBody, http.StatusBadRequest, err)
		return
	}

	if req.Email == nil && req.Name == nil && req.Role == nil {
		writeAppError(w, apperrors.CodeValidationNoFields, http.StatusBadRequest, errors.New("at least one field is required"))
		return
	}

	var role *models.Role
	if req.Role != nil {
		parsedRole := models.Role(*req.Role)
		if !parsedRole.IsValid() {
			writeAppError(w, apperrors.CodeValidationInvalidRole, http.StatusBadRequest, errors.New("role must be ADMIN or USER"))
			return
		}
		role = &parsedRole
	}

	user, err := h.users.Update(r.Context(), id, repository.UpdateUserInput{
		Email: req.Email,
		Name:  req.Name,
		Role:  role,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeAppError(w, apperrors.CodeUserNotFound, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, repository.ErrEmailTaken) {
			writeAppError(w, apperrors.CodeUserEmailTaken, http.StatusConflict, err)
			return
		}
		if errors.Is(err, repository.ErrLastAdmin) {
			writeAppError(w, apperrors.CodeUserLastAdmin, http.StatusConflict, err)
			return
		}
		writeAppError(w, apperrors.CodeUserUpdateFailed, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r.Context()) {
		writeAppError(w, apperrors.CodeForbidden, http.StatusForbidden, errors.New("admin access required"))
		return
	}

	id := chi.URLParam(r, "id")

	if err := h.users.SoftDelete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeAppError(w, apperrors.CodeUserNotFound, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, repository.ErrLastAdmin) {
			writeAppError(w, apperrors.CodeUserLastAdmin, http.StatusConflict, err)
			return
		}
		writeAppError(w, apperrors.CodeUserDeleteFailed, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Me godoc
//
//	@Summary		Get current user
//	@Description	Return the authenticated user's profile
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	models.User
//	@Failure		401	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Router			/api/users/me [get]
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeAppError(w, apperrors.CodeUnauthorized, http.StatusUnauthorized, errors.New("user id not found in context"))
		return
	}

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeAppError(w, apperrors.CodeUserNotFound, http.StatusNotFound, err)
			return
		}
		writeAppError(w, apperrors.CodeUserGetFailed, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func canReadUser(r *http.Request, targetID string) bool {
	if auth.IsAdmin(r.Context()) {
		return true
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	return ok && userID == targetID
}
