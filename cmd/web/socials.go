package main

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"textonly.islandwind.me/internal/data"
)

type SocialResponse struct {
	Metadata data.Metadata `json:"metadata"`
	Data     data.Social   `json:"data"`
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
