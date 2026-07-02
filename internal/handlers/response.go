package handlers

import (
	"encoding/json"
	"net/http"

	"study-tracker-backend/internal/apperrors"
)

type errorResponse = apperrors.Response

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeAppError(w http.ResponseWriter, code apperrors.Code, status int, err error) {
	apperrors.Write(w, apperrors.New(code, status, err))
}
