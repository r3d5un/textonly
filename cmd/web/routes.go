package main

import (
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/home/", app.home)
	mux.HandleFunc("/post/read/", app.readPost)
	mux.HandleFunc("/post/", app.posts)
	mux.HandleFunc("/about/", app.about)
	mux.HandleFunc("/feed.rss", app.feed)

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(mux)
}
