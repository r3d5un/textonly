package main

import (
	"log/slog"
	"net/http"

	"github.com/justinas/alice"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "textonly.islandwind.me/docs"
	"textonly.islandwind.me/ui"
)

func (app *application) routes() http.Handler {
	slog.Info("creating multiplexer")
	mux := http.NewServeMux()

	slog.Info("creating middleware chains")
	// for routes that require authentication
	protected := alice.New(app.basicAuth)

	// static files
	fileServer := http.FileServer(http.FS(ui.Files))
	mux.Handle("GET /static/{filepath...}", fileServer)

	// healthcheck
	slog.Info("adding healthcheck route")
	mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)

	// swagger
	slog.Info("adding swagger documentation routes")
	mux.HandleFunc("GET /swagger/{any...}", httpSwagger.WrapHandler)

	// UI
	mux.HandleFunc("GET /", app.home)
	mux.HandleFunc("GET /home", app.home)
	mux.HandleFunc("GET /post/read/{id}", app.readPost)
	mux.HandleFunc("GET /post", app.posts)
	mux.HandleFunc("GET /about", app.about)
	mux.HandleFunc("GET /feed.rss", app.feed)

	// API
	mux.HandleFunc("GET /api/post", app.listBlogHandler)
	mux.HandleFunc("GET /api/post/{id}", app.getBlogHandler)
	mux.Handle("POST /api/post", protected.ThenFunc(app.postBlogHandler))
	mux.Handle("DELETE /api/post/{id}", protected.ThenFunc(app.deleteBlogHandler))
	mux.Handle("PUT /api/post", protected.ThenFunc(app.updateBlogHandler))

	mux.HandleFunc("GET /api/social", app.listSocialHandler)
	mux.HandleFunc("GET /api/social/{id}", app.getSocialHandler)
	mux.Handle("POST /api/social", protected.ThenFunc(app.postSocialHandler))
	mux.Handle("DELETE /api/social/{id}", protected.ThenFunc(app.deleteSocialHandler))
	mux.Handle("PUT /api/social", protected.ThenFunc(app.updateSocialHandler))

	mux.HandleFunc("GET /api/user/{id}", app.getUserHandler)
	mux.Handle("PUT /api/user", protected.ThenFunc(app.updateUserHandler))

	standard := alice.New(app.recoverPanic, app.rateLimit, app.logRequest, secureHeaders)

	return standard.Then(mux)
}
