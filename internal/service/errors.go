package service

import "net/http"

type AppError struct {
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	Details    []string `json:"details,omitempty"`
	HTTPStatus int      `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

func ValidationError(msg string, details ...string) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: msg, Details: details, HTTPStatus: http.StatusBadRequest}
}
func NotFoundError(msg string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: msg, HTTPStatus: http.StatusNotFound}
}
func ConflictError(msg string) *AppError {
	return &AppError{Code: "CONFLICT", Message: msg, HTTPStatus: http.StatusConflict}
}
func InternalError(msg string) *AppError {
	return &AppError{Code: "INTERNAL_ERROR", Message: msg, HTTPStatus: http.StatusInternalServerError}
}
