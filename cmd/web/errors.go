package main

import (
	"fmt"
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
		app.logger.Error("an error occurred while returning error response", "error", err)
		w.WriteHeader(status)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.ErrorContext(r.Context(), "an unexpected error occurred", "request", r, "error", err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, message string) {
	app.logger.InfoContext(r.Context(), "retirning bad request response", "request", r)
	app.errorResponse(w, r, http.StatusBadRequest, message)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.logger.InfoContext(r.Context(), "returning not found response", "request", r)
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	app.logger.InfoContext(r.Context(), "returning method not allowed response", "request", r)
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}
