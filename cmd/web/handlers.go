package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"textonly.islandwind.me/internal/data"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.redirectToLatestPost(w, r)
}

func (app *application) readPost(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	app.logger.Info("querying blogpost", "id", id)
	blogPost, err := app.models.BlogPosts.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.logger.Info("retrieved post", "id", blogPost.ID, "title", blogPost.Title)

	app.render(w, http.StatusOK, "read.tmpl", &templateData{
		BlogPost: blogPost,
	})
}

func (app *application) posts(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("querying blogposts")
	blogPosts, err := app.models.BlogPosts.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.logger.Info("retrieved blogposts", "number", len(blogPosts))

	app.render(w, http.StatusOK, "posts.tmpl", &templateData{
		BlogPosts: blogPosts,
	})
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("querying user data")
	user, err := app.models.Users.Get(1)
	if err != nil {
		app.logger.Error("unable to query user data", "error", err)
		app.serverError(w, err)
		return
	}
	app.logger.Info("retrieved user data", "user", user)

	app.logger.Info("querying social data")
	socials, err := app.models.Socials.GetByUserID(user.ID)
	if err != nil {
		app.logger.Error("uanble to query social data", "error", err)
		app.serverError(w, err)
		return
	}
	app.logger.Info("retrieved socials", "socials", socials)

	app.render(w, http.StatusOK, "about.tmpl", &templateData{
		User:    user,
		Socials: socials,
	})
}

func (app *application) feed(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("querying blogposts")
	blogPosts, err := app.models.BlogPosts.GetAll()
	if err != nil {
		app.logger.Error("unable to query blogposts", "error", err)
		app.serverError(w, err)
		return
	}
	app.logger.Info("retrieved blogposts", "amount", len(blogPosts))

	app.renderXML(w, http.StatusOK, &templateData{
		BlogPosts: blogPosts,
	})
}
