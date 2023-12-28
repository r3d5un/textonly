package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	app.logger.Error("a server error occurred", "error", err, "trace", debug.Stack())

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (app *application) renderXML(w http.ResponseWriter, status int, data *templateData) {
	ts, ok := app.templateCache["feed"]
	if !ok {
		err := fmt.Errorf("the feed template does not exist")
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "feed.tmpl", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

func GetURL() string {
	url, err := os.LookupEnv("URL")
	if !err {
		url = ":8080"
	}

	return url
}

func GetDBDSN() string {
	dsn, err := os.LookupEnv("DSN")
	if !err {
		dsn = "user=postgres " +
			"password=postgres " +
			"host=localhost " +
			"port=5432 " +
			"dbname=blog " +
			"sslmode=disable "
	}

	return dsn
}

func (app *application) writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
	headers http.Header,
) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) redirectToLatestPost(w http.ResponseWriter, r *http.Request) {
	posts, err := app.models.BlogPosts.LastN(1)
	if err != nil {
		app.logger.Error("error occurred while redirecting", "error", err)
		return
	}

	if len(posts) != 1 {
		app.logger.Error("unexected number of posts returned", "posts", posts)
		app.notFound(w)
	}

	urlString := fmt.Sprintf("/post/read/%d", posts[0].ID)

	app.logger.Info("redirecting to last post", "url", urlString)
	http.Redirect(w, r, urlString, http.StatusMovedPermanently)
}
