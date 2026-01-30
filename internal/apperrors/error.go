package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code   Code
	Status int
	Err    error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Err)
	}
	return string(e.Code)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code Code, status int, err error) *AppError {
	return &AppError{
		Code:   code,
		Status: status,
		Err:    err,
	}
}

func As(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}

func From(err error) *AppError {
	if appErr, ok := As(err); ok {
		return appErr
	}
	return New(CodeInternalError, http.StatusInternalServerError, err)
}
