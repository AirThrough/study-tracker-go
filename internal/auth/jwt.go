package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "userID"

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type Service struct {
	secret []byte
}

func NewService(secret string) *Service {
	return &Service{secret: []byte(secret)}
}

func (s *Service) CreateToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func (s *Service) ParseToken(tokenString string) (Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return Claims{}, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return *claims, nil
}

func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := s.ParseToken(tokenString)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
