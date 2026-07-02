package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"study-tracker-backend/internal/apperrors"
	"study-tracker-backend/internal/auth"
	"study-tracker-backend/internal/repository"
)

type AuthHandler struct {
	users         *repository.UserRepository
	refreshTokens *repository.RefreshTokenRepository
	auth          *auth.Service
	cookieSecure  bool
}

func NewAuthHandler(
	users *repository.UserRepository,
	refreshTokens *repository.RefreshTokenRepository,
	authService *auth.Service,
	cookieSecure bool,
) *AuthHandler {
	return &AuthHandler{
		users:         users,
		refreshTokens: refreshTokens,
		auth:          authService,
		cookieSecure:  cookieSecure,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

// Login godoc
//
//	@Summary		Login
//	@Description	Authenticate and receive a JWT access token. A refresh token is set as an HttpOnly cookie.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginRequest	true	"Credentials"
//	@Success		200		{object}	authResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		401		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAppError(w, apperrors.CodeInvalidRequestBody, http.StatusBadRequest, err)
		return
	}

	if req.Email == "" || req.Password == "" {
		writeAppError(w, apperrors.CodeValidationRequired, http.StatusBadRequest, errors.New("email and password are required"))
		return
	}

	user, passwordHash, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeAppError(w, apperrors.CodeAuthInvalidCredentials, http.StatusUnauthorized, err)
			return
		}
		writeAppError(w, apperrors.CodeAuthenticateFailed, http.StatusInternalServerError, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		writeAppError(w, apperrors.CodeAuthInvalidCredentials, http.StatusUnauthorized, err)
		return
	}

	token, err := h.auth.CreateToken(user.ID, user.Role)
	if err != nil {
		writeAppError(w, apperrors.CodeTokenCreateFailed, http.StatusInternalServerError, err)
		return
	}

	rawRefresh, refreshHash, err := auth.GenerateRefreshToken()
	if err != nil {
		writeAppError(w, apperrors.CodeRefreshTokenCreateFailed, http.StatusInternalServerError, err)
		return
	}

	expiresAt := time.Now().UTC().Add(auth.RefreshTokenTTL)
	if err := h.refreshTokens.Create(r.Context(), user.ID, refreshHash, expiresAt); err != nil {
		writeAppError(w, apperrors.CodeRefreshTokenStoreFailed, http.StatusInternalServerError, err)
		return
	}

	auth.SetRefreshTokenCookie(w, rawRefresh, h.cookieSecure)
	writeJSON(w, http.StatusOK, authResponse{Token: token})
}

// Refresh godoc
//
//	@Summary		Refresh access token
//	@Description	Exchange the refresh token cookie for a new access token
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	authResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/api/auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.RefreshTokenCookie)
	if err != nil || cookie.Value == "" {
		writeAppError(w, apperrors.CodeAuthRefreshMissing, http.StatusUnauthorized, errors.New("refresh token cookie is missing"))
		return
	}

	tokenHash := auth.HashRefreshToken(cookie.Value)
	userID, err := h.refreshTokens.GetValid(r.Context(), tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) {
			auth.ClearRefreshTokenCookie(w, h.cookieSecure)
			writeAppError(w, apperrors.CodeAuthRefreshInvalid, http.StatusUnauthorized, err)
			return
		}
		writeAppError(w, apperrors.CodeRefreshTokenValidateFailed, http.StatusInternalServerError, err)
		return
	}

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			auth.ClearRefreshTokenCookie(w, h.cookieSecure)
			_ = h.refreshTokens.Delete(r.Context(), tokenHash)
			writeAppError(w, apperrors.CodeAuthRefreshInvalid, http.StatusUnauthorized, err)
			return
		}
		writeAppError(w, apperrors.CodeRefreshFailed, http.StatusInternalServerError, err)
		return
	}

	if err := h.refreshTokens.Delete(r.Context(), tokenHash); err != nil {
		writeAppError(w, apperrors.CodeRefreshTokenRotateFailed, http.StatusInternalServerError, err)
		return
	}

	rawRefresh, refreshHash, err := auth.GenerateRefreshToken()
	if err != nil {
		writeAppError(w, apperrors.CodeRefreshTokenCreateFailed, http.StatusInternalServerError, err)
		return
	}

	expiresAt := time.Now().UTC().Add(auth.RefreshTokenTTL)
	if err := h.refreshTokens.Create(r.Context(), user.ID, refreshHash, expiresAt); err != nil {
		writeAppError(w, apperrors.CodeRefreshTokenStoreFailed, http.StatusInternalServerError, err)
		return
	}

	token, err := h.auth.CreateToken(user.ID, user.Role)
	if err != nil {
		writeAppError(w, apperrors.CodeTokenCreateFailed, http.StatusInternalServerError, err)
		return
	}

	auth.SetRefreshTokenCookie(w, rawRefresh, h.cookieSecure)
	writeJSON(w, http.StatusOK, authResponse{Token: token})
}

// Logout godoc
//
//	@Summary		Logout
//	@Description	Revoke the refresh token and clear the refresh token cookie
//	@Tags			auth
//	@Success		204
//	@Router			/api/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.RefreshTokenCookie); err == nil && cookie.Value != "" {
		_ = h.refreshTokens.Delete(r.Context(), auth.HashRefreshToken(cookie.Value))
	}

	auth.ClearRefreshTokenCookie(w, h.cookieSecure)
	w.WriteHeader(http.StatusNoContent)
}
