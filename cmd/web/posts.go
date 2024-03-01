package main

import (
	"errors"
	"net/http"
	"strconv"

	"textonly.islandwind.me/internal/data"
	"textonly.islandwind.me/internal/utils"
	"textonly.islandwind.me/internal/validator"
)

type BlogPostResponse struct {
	Metadata data.Metadata `json:"metadata"`
	Data     data.BlogPost `json:"data"`
}

type BlogPostListResponse struct {
	Metadata data.Metadata    `json:"metadata"`
	Data     []*data.BlogPost `json:"data"`
}

type BlogPostRequest struct {
	Title string `json:"title"`
	Lead  string `json:"lead"`
	Post  string `json:"post_content"`
}

type UpdateBlogResponse struct {
	Message      string `json:"message,omitempty"`
	ID           int    `json:"id,omitempty"`
	RowsAffected int64  `json:"rows_affected,omitempty"`
}

// @Summary		Get a blog post
// @Description	Get a blog post by ID
// @Param			id	path	string	true	"ID (int)"
// @Tags			Blog Post
//
// @Produce		json
// @Success		200	{object}	BlogPostResponse
// @Failure		500	{object}	ErrorMessage
// @Failure		401	{object}	ErrorMessage
// @Failure		404	{object}	ErrorMessage
// @Failure		429	{object}	ErrorMessage
// @Router			/api/post/{id} [get]
func (app *application) getBlogHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	logger.InfoContext(ctx, "parsing blog ID from path", "key", "id", "path", r.URL.Path)
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
		app.notFound(w)
		return
	}

	logger.InfoContext(ctx, "retrieving post", "id", id)
	bp, err := app.models.BlogPosts.Get(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			logger.InfoContext(ctx, "no records found", "id", id)
			app.notFoundResponse(w, r)
			return
		default:
			logger.ErrorContext(ctx, "an error occurred during retrieval", "error", err)
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	logger.InfoContext(ctx, "returning blog post", "id", bp.ID, "title", bp.Title)
	err = app.writeJSON(
		w,
		http.StatusOK,
		BlogPostResponse{Metadata: data.Metadata{}, Data: *bp},
		nil,
	)
	if err != nil {
		logger.ErrorContext(ctx, "unable to write response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

// @Summary		List blog posts
// @Description	List blog posts
// @Tags			Blog Post
// @Produce		json
// @Param			id					query		int		false	"id"
// @Param			title				query		string	false	"title"
// @Param			lead				query		string	false	"lead"
// @Param			created_from		query		string	false	"created_from"
// @Param			created_to			query		string	false	"created_to"
// @Param			last_updated_from	query		string	false	"last_updated_from"
// @Param			last_updated_to		query		string	false	"last_updated_to"
// @Param			order_by			query		string	false	"order_by"
//
// @Success		200					{object}	BlogPostListResponse
//
// @Failure		500					{object}	ErrorMessage
// @Failure		401					{object}	ErrorMessage
//
// @Failure		404					{object}	ErrorMessage
// @Failure		429					{object}	ErrorMessage
// @Router			/api/post/ [get]
func (app *application) listBlogHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	var input struct {
		data.Filters `json:"filters,omitempty"`
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Filters.ID = app.readQueryParamToIntPtr(qs, "id", v)
	input.Filters.Title = app.readQueryString(qs, "title", "")
	input.Filters.Lead = app.readQueryString(qs, "lead", "")
	input.Filters.Post = app.readQueryString(qs, "post", "")
	input.Filters.CreatedFrom = app.readQueryDate(qs, "created_from", v)
	input.Filters.CreatedTo = app.readQueryDate(qs, "created_to", v)
	input.Filters.LastUpdatedFrom = app.readQueryDate(qs, "last_updated_from", v)
	input.Filters.LastUpdatedTo = app.readQueryDate(qs, "last_updated_to", v)

	input.Filters.Page = app.readQueryInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readQueryInt(qs, "page_size", 50_000, v)

	input.Filters.OrderBy = app.readQueryCommaSeperatedString(qs, "order_by", "-last_updated_from")
	input.Filters.OrderBySafeList = []string{
		"lead", "title", "post", "created_from", "created_to", "last_updated_from, last_updated_to",
		"-lead", "-title", "-post", "-created_from", "-created_to", "-last_updated_from", "-last_updated_to",
		"id", "-id",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	bp, metadata, err := app.models.BlogPosts.GetAll(ctx, input.Filters)
	if err != nil {
		logger.ErrorContext(ctx, "unable to get all blog posts", "error", err, "input", input)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(
		w, http.StatusOK, BlogPostListResponse{Metadata: metadata, Data: bp}, nil,
	)
	if err != nil {
		logger.ErrorContext(ctx, "error writing response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

// @Summary		Post a blog post
// @Description	Post a blog post by ID
//
// @Param			BlogPostRequest	body	BlogPostRequest	true	"Push Blog Post"
//
// @Tags			Blog Post
// @Produce		json
// @Success		200	{object}	BlogPostResponse
// @Failure		500	{object}	ErrorMessage
//
// @Failure		401	{object}	ErrorMessage
//
// @Failure		404	{object}	ErrorMessage
// @Failure		429	{object}	ErrorMessage
// @Router			/api/post/{id} [post]
func (app *application) postBlogHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	var blogPost BlogPostRequest

	err := app.readJSON(r, &blogPost)
	if err != nil {
		logger.ErrorContext(ctx, "unable to parse JSON request body", "error", err)
		app.badRequestResponse(w, r, "unable to parse JSON request body")
		return
	}

	bp, err := app.models.BlogPosts.Insert(ctx, &data.BlogPost{
		Title: blogPost.Title,
		Lead:  blogPost.Lead,
		Post:  blogPost.Post,
	})
	if err != nil {
		logger.ErrorContext(ctx, "unable to create blog post", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, bp, nil)
	if err != nil {
		logger.ErrorContext(ctx, "unable to write response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

// @Summary		Delete a blog post
// @Description	Delete a blog post by ID
// @Param			id	path	string	true	"ID (int)"
// @Tags			Blog Post
// @Produce		json
// @Success		200	{object}	UpdateBlogResponse
// @Failure		500	{object}	ErrorMessage
// @Failure		401	{object}	ErrorMessage
// @Failure		404	{object}	ErrorMessage
// @Failure		429	{object}	ErrorMessage
// @Router			/api/post/{id} [delete]
func (app *application) deleteBlogHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	logger.InfoContext(ctx, "parsing blog ID from path", "key", "id")
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
		app.notFound(w)
		return
	}

	rowsAffected, err := app.models.BlogPosts.Delete(ctx, id)
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
		UpdateBlogResponse{Message: "blog post deleted", RowsAffected: rowsAffected, ID: id},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// @Summary		Update a blog post
// @Description	Update a blog post by ID
//
// @Param			data.BlogPost	body	data.BlogPost	true	"Update Blog Post"
//
// @Tags			Blog Post
//
// @Produce		json
// @Success		200	{object}	UpdateBlogResponse
// @Failure		500	{object}	ErrorMessage
// @Failure		401	{object}	ErrorMessage
// @Failure		404	{object}	ErrorMessage
// @Failure		429	{object}	ErrorMessage
// @Router			/api/post/{id} [put]
func (app *application) updateBlogHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	var input data.BlogPost

	err := app.readJSON(r, &input)
	if err != nil {
		logger.Error(
			"unable to parse JSON request body",
			"error", err,
			"request", r.Body,
		)
		app.badRequestResponse(w, r, "unable to parse JSON request body")
		return
	}

	rowsAffected, err := app.models.BlogPosts.Update(ctx, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(
		w,
		http.StatusOK,
		UpdateBlogResponse{Message: "blog post updated", RowsAffected: rowsAffected, ID: input.ID},
		nil,
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
