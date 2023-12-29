package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"textonly.islandwind.me/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	// healthcheck
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// UI
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/home", app.home)
	router.HandlerFunc(http.MethodGet, "/post/read/:id", app.readPost)
	router.HandlerFunc(http.MethodGet, "/post", app.posts)
	router.HandlerFunc(http.MethodGet, "/about", app.about)
	router.HandlerFunc(http.MethodGet, "/feed.rss", app.feed)

	// API
	// TODO: Protect API endspoints
	router.HandlerFunc(http.MethodGet, "/api/post", app.listBlogHandler)
	router.HandlerFunc(http.MethodGet, "/api/post/:id", app.getBlogHandler)
	router.HandlerFunc(http.MethodPost, "/api/post", app.postBlogHandler)

	standard := alice.New(app.recoverPanic, app.rateLimit, app.logRequest, secureHeaders)

	return standard.Then(router)
}
