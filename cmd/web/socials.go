package main

import (
	"errors"
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

// @Summary		Get social data
// @Description	Get social data by ID
// @Param			id	path	string	true	"ID (int)"
// @Tags			Social
// @Produce		json
// @Success		200	{object}	SocialResponse
// @Failure		500	{object}	ErrorMessage
// @Failure		401	{object}	ErrorMessage
// @Failure		404	{object}	ErrorMessage
// @Failure		429	{object}	ErrorMessage
// @Router			/api/social/{id} [get]
func (app *application) getSocialHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	app.logger.InfoContext(ctx, "parsing social ID from path", "key", "id", "path", r.URL.Path)
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.logger.ErrorContext(
			ctx,
			"unable to get ID parameter from URL string",
			"params",
			params,
			"error",
			err,
		)
		app.serverErrorResponse(w, r, err)
		return
	}
	if id < 0 {
		app.logger.InfoContext(ctx, "invalid ID", "id", id)
		app.notFoundResponse(w, r)
		return
	}

	app.logger.InfoContext(ctx, "retrieving social account data", "id", id)
	s, err := app.models.Socials.Get(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.logger.InfoContext(ctx, "no records found", "id", id)
			app.notFoundResponse(w, r)
			return
		default:
			app.logger.ErrorContext(ctx, "an error occurred during retrieval", "error", err)
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	app.logger.InfoContext(ctx, "returning social account data", "id", s.ID)
	err = app.writeJSON(
		w,
		http.StatusOK,
		SocialResponse{Metadata: data.Metadata{}, Data: *s},
		nil,
	)
	if err != nil {
		app.logger.ErrorContext(ctx, "unable to write response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

// @Summary		List social data
// @Description	List social data
// @Tags			Social
// @Produce		json
// @Param			id				query		int		false	"id"
// @Param			user_id			query		int		false	"user_id"
// @Param			social_platform	query		string	false	"social_platform"
// @Success		200				{object}	SocialListResponse
//
// @Failure		500				{object}	ErrorMessage
// @Failure		401				{object}	ErrorMessage
//
// @Failure		404				{object}	ErrorMessage
// @Failure		429				{object}	ErrorMessage
// @Router			/api/social/ [get]
func (app *application) listSocialHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var input struct {
		data.Filters `json:"filters,omitempty"`
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Filters.ID = app.readQueryParamToIntPtr(qs, "id", v)
	input.Filters.UserID = app.readQueryParamToIntPtr(qs, "user_id", v)
	input.Filters.SocialPlatform = app.readQueryString(qs, "social_platform", "")

	input.Filters.Page = app.readQueryInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readQueryInt(qs, "page_size", 50_000, v)

	input.Filters.OrderBy = app.readQueryCommaSeperatedString(qs, "order_by", "-id")
	input.Filters.OrderBySafeList = []string{
		"id", "user_id", "social_platform",
		"-id", "-user_id", "-social_platform",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	ss, metadata, err := app.models.Socials.GetAll(ctx, input.Filters)
	if err != nil {
		app.logger.ErrorContext(ctx, "unable to get social data", "error", err, "input", input)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(
		w, http.StatusOK, SocialListResponse{Metadata: metadata, Data: ss}, nil,
	)
	if err != nil {
		app.logger.ErrorContext(ctx, "error writing response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

// @Summary		Post social data
// @Description	Post social data
//
// @Param			SocialPostRequest	body	SocialPostRequest	true	"Push social data"
//
// @Tags			Social
// @Produce		json
// @Success		200	{object}	SocialResponse
// @Failure		500	{object}	ErrorMessage
//
// @Failure		401	{object}	ErrorMessage
//
// @Failure		404	{object}	ErrorMessage
// @Failure		429	{object}	ErrorMessage
// @Router			/api/social/{id} [post]
func (app *application) postSocialHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var s SocialPostRequest

	err := app.readJSON(r, &s)
	if err != nil {
		app.logger.ErrorContext(ctx, "unable to parse JSON request body", "error", err)
		app.badRequestResponse(w, r, "unable to parse JSON request body")
		return
	}

	queryResponse, err := app.models.Socials.Insert(ctx, &data.Social{
		UserID:         s.UserID,
		SocialPlatform: s.SocialPlatform,
		Link:           s.Link,
	})
	if err != nil {
		app.logger.ErrorContext(ctx, "unable to create social data", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, queryResponse, nil)
	if err != nil {
		app.logger.ErrorContext(ctx, "unable to write response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

// @Summary		Update social data
// @Description	Update social data
//
// @Param			data.Social	body	data.Social	true	"Update Social Data"
//
// @Tags			Social
//
// @Produce		json
// @Success		200	{object}	UpdateSocialResponse
// @Failure		500	{object}	ErrorMessage
// @Failure		401	{object}	ErrorMessage
// @Failure		404	{object}	ErrorMessage
// @Failure		429	{object}	ErrorMessage
// @Router			/api/social/{id} [put]
func (app *application) updateSocialHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input data.Social

	err := app.readJSON(r, &input)
	if err != nil {
		app.logger.ErrorContext(
			ctx,
			"unable to parse JSON request body",
			"error",
			err,
			"request",
			r.Body,
		)
		app.badRequestResponse(w, r, "uanble to parse JSON request body")
		return
	}

	rowsAffected, err := app.models.Socials.Update(ctx, &input)
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

// @Summary		Delete social data
// @Description	Delete social data
// @Param			id	path	string	true	"ID (int)"
// @Tags			Social
// @Produce		json
// @Success		200	{object}	UpdateSocialResponse
// @Failure		500	{object}	ErrorMessage
// @Failure		401	{object}	ErrorMessage
// @Failure		404	{object}	ErrorMessage
// @Failure		429	{object}	ErrorMessage
// @Router			/api/social/{id} [delete]
func (app *application) deleteSocialHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	app.logger.InfoContext(ctx, "parsing social ID from path", "key", "id", "path", r.URL.Path)
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.logger.ErrorContext(
			ctx,
			"unable to get ID parameter from URL string",
			"params",
			params,
			"error",
			err,
		)
		app.serverErrorResponse(w, r, err)
		return
	}
	if id < 0 {
		app.logger.InfoContext(ctx, "invalid ID", "id", id)
		app.notFound(w)
		return
	}

	rowsAffected, err := app.models.Socials.Delete(ctx, id)
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
