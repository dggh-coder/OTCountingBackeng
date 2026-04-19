package api

import (
	"encoding/json"
	"net/http"
	"otcountingbackend/internal/service"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*service.AppError); ok {
		writeJSON(w, appErr.HTTPStatus, appErr)
		return
	}
	writeJSON(w, http.StatusInternalServerError, &service.AppError{Code: "INTERNAL_ERROR", Message: "unexpected error"})
}
