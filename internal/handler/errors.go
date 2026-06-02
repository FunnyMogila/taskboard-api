package handler

import (
	"errors"
	"net/http"

	"taskboard-api/internal/errs"
)

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errs.ErrNotFound):
		writeError(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, errs.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "resource already exists")
	case errors.Is(err, errs.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, errs.ErrProjectClosed):
		writeError(w, http.StatusConflict, "project is closed")
	case errors.Is(err, errs.ErrInvalidTransition):
		writeError(w, http.StatusConflict, "invalid status transition")
	case errors.Is(err, errs.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
