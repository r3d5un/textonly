package main

import (
	"errors"
	"net/http"
	"strconv"

	"textonly.islandwind.me/internal/data"
	"textonly.islandwind.me/internal/utils"
	"textonly.islandwind.me/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	logger.InfoContext(r.Context(), "redirecting to latest post", "request", r.URL)
	app.redirectToLatestPost(w, r)
}

func (app *application) readPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	logger.InfoContext(ctx, "parsing id parameter")
	rawValue := r.PathValue("id")
	if rawValue == "" {
		logger.ErrorContext(ctx, "parameter value empty", "id", rawValue)
		app.badRequestResponse(w, r, "parameter value empty")
		return
	}
	logger.InfoContext(ctx, "parsing id value", "id", rawValue)

	logger.InfoContext(ctx, "parsing id value to integer", "id", rawValue)
	id, err := strconv.Atoi(rawValue)
	if err != nil {
		logger.ErrorContext(ctx, "unable to parse id value", "value", rawValue)
		app.badRequestResponse(w, r, "unable to parse id value")
		return
	}
	if err != nil || id < 1 {
		logger.ErrorContext(ctx, "invalid id value", "value", id)
		app.notFound(w)
		return
	}
	logger.ErrorContext(ctx, "id value parsed", "id", id)

	logger.InfoContext(ctx, "querying blogpost", "id", id)
	blogPost, err := app.models.BlogPosts.Get(ctx, id)
	if err != nil {
		if errors.Is(err, data.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	logger.InfoContext(ctx, "retrieved post", "id", blogPost.ID, "title", blogPost.Title)

	app.render(ctx, w, http.StatusOK, "read.tmpl", &templateData{
		BlogPost: blogPost,
	})
}

func (app *application) posts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	var input struct {
		data.Filters `json:"filters,omitempty"`
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Filters.Page = app.readQueryInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readQueryInt(qs, "page_size", 50_000, v)

	logger.InfoContext(ctx, "querying blogposts")
	blogPosts, _, err := app.models.BlogPosts.GetAll(ctx, input.Filters)
	if err != nil {
		app.serverError(w, err)
		return
	}
	logger.InfoContext(ctx, "retrieved blogposts", "number", len(blogPosts))

	app.render(ctx, w, http.StatusOK, "posts.tmpl", &templateData{
		BlogPosts: blogPosts,
	})
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	logger.InfoContext(ctx, "querying user data", "id", 1)
	user, err := app.models.Users.Get(ctx, 1)
	if err != nil {
		logger.ErrorContext(ctx, "unable to query user data", "error", err)
		app.serverError(w, err)
		return
	}
	logger.InfoContext(ctx, "retrieved user data", "user", user)

	logger.Info("querying social data")
	socials, err := app.models.Socials.GetByUserID(ctx, user.ID)
	if err != nil {
		logger.ErrorContext(ctx, "uanble to query social data", "error", err)
		app.serverError(w, err)
		return
	}
	logger.InfoContext(ctx, "retrieved socials", "socials", socials)

	app.render(ctx, w, http.StatusOK, "about.tmpl", &templateData{
		User:    user,
		Socials: socials,
	})
}

func (app *application) feed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := utils.LoggerFromContext(ctx)

	var input struct {
		data.Filters `json:"filters,omitempty"`
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Filters.Page = app.readQueryInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readQueryInt(qs, "page_size", 50_000, v)

	logger.InfoContext(ctx, "querying blogposts")
	blogPosts, _, err := app.models.BlogPosts.GetAll(ctx, input.Filters)
	if err != nil {
		logger.ErrorContext(ctx, "unable to query blogposts", "error", err)
		app.serverError(w, err)
		return
	}
	logger.InfoContext(ctx, "retrieved blogposts", "amount", len(blogPosts))

	app.renderXML(ctx, w, http.StatusOK, &templateData{
		BlogPosts: blogPosts,
	})
}
