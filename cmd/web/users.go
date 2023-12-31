package main

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"textonly.islandwind.me/internal/data"
)

type UserPostResponse struct {
	Metadata data.Metadata `json:"metadata"`
	Data     data.User     `json:"data"`
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("parsing user ID from path", "key", "id", "path", r.URL.Path)
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

	slog.Info("retrieving user", "id", id)
	user, err := app.models.Users.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			slog.Info("user not found", "id", id)
			app.notFoundResponse(w, r)
			return
		default:
			slog.Error("an error occurred during retrieval", "error", err)
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	slog.Info("returning user", "id", user.ID, "name", user.Name)
	err = app.writeJSON(
		w,
		http.StatusOK,
		UserPostResponse{Metadata: data.Metadata{}, Data: *user},
		nil,
	)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}
