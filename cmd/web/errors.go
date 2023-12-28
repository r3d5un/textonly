package main

import (
	"log/slog"
	"net/http"
)

type ErrorMessage struct {
	Message any `json:"message"`
}

func (app *application) errorResponse(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	message any,
) {
	err := app.writeJSON(w, status, ErrorMessage{Message: message}, nil)
	if err != nil {
		slog.Error("an error occurred while returning error response", "error", err)
		w.WriteHeader(500)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("an unexpected error occurred", "request", r, "error", err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, message string) {
	app.errorResponse(w, r, http.StatusBadRequest, message)
}
