package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/home", app.home)
	router.HandlerFunc(http.MethodGet, "/post/read/:id", app.readPost)
	router.HandlerFunc(http.MethodGet, "/post", app.posts)
	router.HandlerFunc(http.MethodGet, "/about", app.about)
	router.HandlerFunc(http.MethodGet, "/feed.rss", app.feed)

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
