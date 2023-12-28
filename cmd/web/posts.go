package main

import (
	"log/slog"
	"net/http"
	"time"

	"textonly.islandwind.me/internal/data"
)

type BlogPostRequest struct {
	Title string `json:"title"`
	Lead  string `json:"lead"`
	Post  string `json:"post_content"`
}

type BlogPostResponse struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Lead       string    `json:"lead"`
	Post       string    `json:"post"`
	LastUpdate time.Time `json:"last_update"`
	Created    time.Time `json:"created"`
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
