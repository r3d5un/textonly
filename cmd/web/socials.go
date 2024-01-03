package main

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"textonly.islandwind.me/internal/data"
	"textonly.islandwind.me/internal/validator"
)

type SocialResponse struct {
	Metadata data.Metadata `json:"metadata"`
	Data     data.Social   `json:"data"`
}

type SocialListResponse struct {
	Metadata data.Metadata  `json:"metadata"`
	Data     []*data.Social `json:"data"`
}

type SocialPostRequest struct {
	UserID         int    `json:"user_id"`
	SocialPlatform string `json:"social_platform"`
	Link           string `json:"link"`
}

type UpdateSocialResponse struct {
	Message      string `json:"message,omitempty"`
	ID           int    `json:"id,omitempty"`
	RowsAffected int64  `json:"rows_affected,omitempty"`
}

func (app *application) getSocialHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("parsing social ID from path", "key", "id", "path", r.URL.Path)
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		slog.Error("unable to get ID parameter from URL string", "params", params, "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
	if id < 0 {
		slog.Info("invalid ID", "id", id)
		app.notFoundResponse(w, r)
		return
	}

	slog.Info("retrieving social account data", "id", id)
	s, err := app.models.Socials.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			slog.Info("no records found", "id", id)
			app.notFoundResponse(w, r)
			return
		default:
			slog.Error("an error occurred during retrieval", "error", err)
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	slog.Info("returning social account data", "id", s.ID)
	err = app.writeJSON(
		w,
		http.StatusOK,
		SocialResponse{Metadata: data.Metadata{}, Data: *s},
		nil,
	)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) listSocialHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters `json:"filters,omitempty"`
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Filters.ID = app.readQueryInt(qs, "id", 0, v)
	input.Filters.UserID = app.readQueryInt(qs, "user_id", 0, v)

	input.Filters.Page = app.readQueryInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readQueryInt(qs, "page_size", 50_000, v)

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	ss, metadata, err := app.models.Socials.GetAll(input.Filters)
	if err != nil {
		slog.Error("unable to get social data", "error", err, "input", input)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(
		w, http.StatusOK, SocialListResponse{Metadata: metadata, Data: ss}, nil,
	)
	if err != nil {
		slog.Error("error writing response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) postSocialHandler(w http.ResponseWriter, r *http.Request) {
	var s SocialPostRequest

	err := app.readJSON(r, &s)
	if err != nil {
		slog.Error("unable to parse JSON request body", "error", err)
		app.badRequestResponse(w, r, "unable to parse JSON request body")
		return
	}

	queryResponse, err := app.models.Socials.Insert(&data.Social{
		UserID:         s.UserID,
		SocialPlatform: s.SocialPlatform,
		Link:           s.Link,
	})
	if err != nil {
		slog.Error("unable to create social data", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, queryResponse, nil)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) putSocialHandler(w http.ResponseWriter, r *http.Request) {
	var input data.Social
	err := app.readJSON(r, &input)
	if err != nil {
		slog.Error("unable to parse JSON request body", "error", err, "request", r.Body)
		app.badRequestResponse(w, r, "uanble to parse JSON request body")
		return
	}

	rowsAffected, err := app.models.Socials.Update(&input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(
		w,
		http.StatusOK,
		UpdateSocialResponse{
			Message:      "social info updated",
			RowsAffected: rowsAffected,
			ID:           input.ID,
		},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) deleteSocialHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("parsing social ID from path", "key", "id", "path", r.URL.Path)
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		slog.Error("unable to get ID parameter from URL string", "params", params, "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
	if id < 0 {
		slog.Info("invalid ID", "id", id)
		app.notFound(w)
		return
	}

	rowsAffected, err := app.models.Socials.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	err = app.writeJSON(
		w,
		http.StatusOK,
		UpdateSocialResponse{Message: "social info deleted", RowsAffected: rowsAffected, ID: id},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
