package apperrors

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	OriginalError string `json:"original_error,omitempty"`
}

func Write(w http.ResponseWriter, err error) {
	appErr := From(err)

	body := Response{
		Code:    string(appErr.Code),
		Message: Message(appErr.Code),
	}
	if appErr.Err != nil {
		body.OriginalError = appErr.Err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Status)
	_ = json.NewEncoder(w).Encode(body)
}
