package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"textonly.islandwind.me/ui"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// error handler routes
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// for routes that require authentication
	protected := alice.New(app.basicAuth)

	// static files
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
	router.HandlerFunc(http.MethodGet, "/api/post", app.listBlogHandler)
	router.HandlerFunc(http.MethodGet, "/api/post/:id", app.getBlogHandler)
	router.Handler(http.MethodPost, "/api/post", protected.ThenFunc(app.postBlogHandler))
	router.Handler(http.MethodDelete, "/api/post/:id", protected.ThenFunc(app.deleteBlogHandler))
	router.Handler(http.MethodPut, "/api/post", protected.ThenFunc(app.updateBlogHandler))

	router.HandlerFunc(http.MethodGet, "/api/social", app.listSocialHandler)
	router.HandlerFunc(http.MethodGet, "/api/social/:id", app.getSocialHandler)
	router.Handler(http.MethodPost, "/api/social", protected.ThenFunc(app.postSocialHandler))
	router.Handler(
		http.MethodDelete,
		"/api/social/:id",
		protected.ThenFunc(app.deleteSocialHandler),
	)
	router.Handler(http.MethodPut, "/api/social", protected.ThenFunc(app.putSocialHandler))

	router.HandlerFunc(http.MethodGet, "/api/user/:id", app.getUserHandler)
	router.Handler(http.MethodPut, "/api/user", protected.ThenFunc(app.updateUserHandler))

	standard := alice.New(app.recoverPanic, app.rateLimit, app.logRequest, secureHeaders)

	return standard.Then(router)
}
