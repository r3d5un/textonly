package main

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"textonly.islandwind.me/internal/data"
)

type BlogPostResponse struct {
	Metadata data.Metadata `json:"metadata"`
	Data     data.BlogPost `json:"data"`
}

type BlogPostRequest struct {
	Title string `json:"title"`
	Lead  string `json:"lead"`
	Post  string `json:"post_content"`
}

func (app *application) getBlogHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("parsing event ID from path", "key", "id", "path", r.URL.Path)
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
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
		default:
			slog.Error("an error occurred during retrieval", "error", err)
			app.serverErrorResponse(w, r, err)
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
