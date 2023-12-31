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

func (app *application) getBlogHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("parsing blog ID from path", "key", "id", "path", r.URL.Path)
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

	slog.Info("retrieving post", "id", id)
	bp, err := app.models.BlogPosts.Get(id)
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

	slog.Info("returning blog post", "id", bp.ID, "title", bp.Title)
	err = app.writeJSON(
		w,
		http.StatusOK,
		BlogPostResponse{Metadata: data.Metadata{}, Data: *bp},
		nil,
	)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) listBlogHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters `json:"filters,omitempty"`
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Filters.ID = app.readQueryInt(qs, "id", 0, v)
	input.Filters.Title = app.readQueryString(qs, "id", "")
	input.Filters.Lead = app.readQueryString(qs, "lead", "")
	input.Filters.Post = app.readQueryString(qs, "post", "")
	input.Filters.Created = app.readQueryDate(qs, "created", v)
	input.Filters.LastUpdate = app.readQueryDate(qs, "last_update", v)

	input.Filters.Page = app.readQueryInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readQueryInt(qs, "page_size", 50_000, v)

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	bp, metadata, err := app.models.BlogPosts.GetAll(input.Filters)
	if err != nil {
		slog.Error("unable to get all blog posts", "error", err, "input", input)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(
		w, http.StatusOK, BlogPostListResponse{Metadata: metadata, Data: bp}, nil,
	)
	if err != nil {
		slog.Error("error writing response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) postBlogHandler(w http.ResponseWriter, r *http.Request) {
	var blogPost BlogPostRequest

	err := app.readJSON(r, &blogPost)
	if err != nil {
		slog.Error("unable to parse JSON request body", "error", err)
		app.badRequestResponse(w, r, "unable to parse JSON request body")
		return
	}

	bp, err := app.models.BlogPosts.Insert(&data.BlogPost{
		Title: blogPost.Title,
		Lead:  blogPost.Lead,
		Post:  blogPost.Post,
	})
	if err != nil {
		slog.Error("unable to create blog post", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, bp, nil)
	if err != nil {
		slog.Error("unable to write response", "error", err)
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) deleteBlogHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("parsing blog ID from path", "key", "id", "path", r.URL.Path)
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

	rowsAffected, err := app.models.BlogPosts.Delete(id)
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

func (app *application) updateBlogHandler(w http.ResponseWriter, r *http.Request) {
	var input data.BlogPost
	err := app.readJSON(r, &input)
	if err != nil {
		slog.Error("unable to parse JSON request body", "error", err, "request", r.Body)
		app.badRequestResponse(w, r, "unable to parse JSON request body")
		return
	}

	rowsAffected, err := app.models.BlogPosts.Update(&input)
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
