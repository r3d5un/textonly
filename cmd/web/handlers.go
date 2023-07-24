package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"textonly.islandwind.me/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Printf("querying blogposts")
	blogPosts, err := app.blogPosts.LastN(3)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.infoLog.Printf("retrieved %d blogposts", len(blogPosts))

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

	app.infoLog.Printf("querying for blog post with id %d", id)
	blogPost, err := app.blogPosts.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	app.infoLog.Printf("retrieved post %d - %s", blogPost.ID, blogPost.Title)

	app.render(w, http.StatusOK, "read.tmpl", &templateData{
		BlogPost: blogPost,
	})
}

func (app *application) posts(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Printf("querying blogposts")
	blogPosts, err := app.blogPosts.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.infoLog.Printf("retrieved %d blogposts", len(blogPosts))

	app.render(w, http.StatusOK, "posts.tmpl", &templateData{
		BlogPosts: blogPosts,
	})
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Printf("querying user data")
	user, err := app.user.Get(1)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.infoLog.Printf("querying social data")
	socials, err := app.sosials.GetByUserID(user.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.infoLog.Printf("retrieved %d socials", len(socials))

	app.render(w, http.StatusOK, "about.tmpl", &templateData{
		User:    user,
		Socials: socials,
	})
}

func (app *application) feed(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Printf("querying blogposts")
	blogPosts, err := app.blogPosts.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.infoLog.Printf("retrieved %d blogposts", len(blogPosts))

	app.renderXML(w, http.StatusOK, &templateData{
		BlogPosts: blogPosts,
	})
}
