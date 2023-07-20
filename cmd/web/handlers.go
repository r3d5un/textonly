package main

import (
	"errors"
	"net/http"
	"strconv"

	"textonly.islandwind.me/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

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
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
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
	if r.URL.Path != "/post/" {
		app.notFound(w)
		return
	}

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
	if r.URL.Path != "/about/" {
		app.notFound(w)
		return
	}

	app.render(w, http.StatusOK, "about.tmpl", &templateData{})
}
