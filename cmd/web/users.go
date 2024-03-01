package main

import (
	"errors"
	"net/http"
	"strconv"

	"textonly.islandwind.me/internal/data"
	"textonly.islandwind.me/internal/utils"
)

type UserPostResponse struct {
	Metadata data.Metadata `json:"metadata"`
	Data     data.User     `json:"data"`
}

type UpdateUserResponse struct {
	Message      string `json:"message,omitempty"`
	RowsAffected int64  `json:"rows_affected,omitempty"`
}

// @Summary		Get user data
// @Description	Get user data
// @Param			id	path	string	true	"ID (int)"
// @Tags			User
//
// @Produce		json
// @Success		200	{object}	UserPostResponse
// @Failure		500	{object}	ErrorMessage
// @Failure		401	{object}	ErrorMessage
// @Failure		404	{object}	ErrorMessage
// @Failure		429	{object}	ErrorMessage
// @Router			/api/user/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	logger.InfoContext(ctx, "parsing user ID from path", "key", "id", "path", r.URL.Path)
	rawValue := r.PathValue("id")
	if rawValue == "" {
		logger.ErrorContext(ctx, "parameter value empty", "id", rawValue)
		app.badRequestResponse(w, r, "parameter value empty")
		return
	}

	id, err := strconv.Atoi(rawValue)
	if err != nil {
		logger.ErrorContext(ctx, "unable to parse id value", "value", rawValue)
		app.badRequestResponse(w, r, "unable to parse id value")
		return
	}
	if id < 0 {
		logger.InfoContext(ctx, "invalid ID", "id", id)
		app.notFoundResponse(w, r)
		return
	}

	logger.InfoContext(ctx, "retrieving user", "id", id)
	user, err := app.models.Users.Get(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			logger.InfoContext(ctx, "user not found", "id", id)
			app.notFoundResponse(w, r)
			return
		default:
			logger.ErrorContext(ctx, "an error occurred during retrieval", "error", err)
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	logger.InfoContext(ctx, "returning user", "id", user.ID, "name", user.Name)
	err = app.writeJSON(
		w,
		http.StatusOK,
		UserPostResponse{Metadata: data.Metadata{}, Data: *user},
		nil,
	)
	if err != nil {
		logger.ErrorContext(ctx, "unable to write response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

// @Summary		Update user data
// @Description	Update user data
//
// @Param			data.User	body	data.User	true	"Update User"
//
// @Tags			User
//
// @Produce		json
// @Success		200	{object}	UpdateUserResponse
// @Failure		500	{object}	ErrorMessage
// @Failure		401	{object}	ErrorMessage
// @Failure		404	{object}	ErrorMessage
// @Failure		429	{object}	ErrorMessage
// @Router			/api/user/{id} [put]
func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	var input data.User
	err := app.readJSON(r, &input)
	if err != nil {
		msg := "unable to parse JSON request body"
		logger.ErrorContext(ctx, msg, "error", err, "request", r.Body)
		app.badRequestResponse(w, r, msg)
		return
	}

	rowsAffected, err := app.models.Users.Update(ctx, &input)
	if err != nil {
		logger.ErrorContext(ctx, "unable to update user", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(
		w,
		http.StatusOK,
		UpdateUserResponse{Message: "user updated", RowsAffected: rowsAffected},
		nil,
	)
	if err != nil {
		logger.ErrorContext(ctx, "unable to write response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}
