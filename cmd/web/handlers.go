package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"textonly.islandwind.me/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	blogPosts, err := app.blogPosts.LastN(3)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, http.StatusOK, "home.tmpl", &templateData{
		BlogPosts: blogPosts,
	})
}

func (app *application) readPost(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	blogPost, err := app.blogPosts.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.render(w, http.StatusOK, "read.tmpl", &templateData{
		BlogPost: blogPost,
	})
}

func (app *application) posts(w http.ResponseWriter, r *http.Request) {
	blogPosts, err := app.blogPosts.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, http.StatusOK, "posts.tmpl", &templateData{
		BlogPosts: blogPosts,
	})
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "about.tmpl", &templateData{})
}

func (app *application) feed(w http.ResponseWriter, r *http.Request) {
	blogPosts, err := app.blogPosts.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.renderXML(w, http.StatusOK, &templateData{
		BlogPosts: blogPosts,
	})
}
