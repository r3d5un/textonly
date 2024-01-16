package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

	"textonly.islandwind.me/internal/validator"
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

func (app *application) render(
	ctx context.Context,
	w http.ResponseWriter,
	status int,
	page string,
	data *templateData,
) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.logger.ErrorContext(ctx, "error occurred while rendering template", "error", err)
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

func (app *application) renderXML(
	ctx context.Context,
	w http.ResponseWriter,
	status int,
	data *templateData,
) {
	ts, ok := app.templateCache["feed"]
	if !ok {
		err := fmt.Errorf("the feed template does not exist")
		app.logger.ErrorContext(ctx, "error occurred while rendering template", "error", err)
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

func (app *application) readJSON(r *http.Request, data interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(data)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) redirectToLatestPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	posts, err := app.models.BlogPosts.LastN(1)
	if err != nil {
		app.logger.ErrorContext(ctx, "error occurred while redirecting", "error", err)
		return
	}

	if len(posts) != 1 {
		app.logger.ErrorContext(ctx, "unexected number of posts returned", "posts", posts)
		app.notFound(w)
	}

	urlString := fmt.Sprintf("/post/read/%d", posts[0].ID)

	app.logger.InfoContext(ctx, "redirecting to last post", "url", urlString)
	http.Redirect(w, r, urlString, http.StatusMovedPermanently)
}

func (app *application) readQueryString(
	qs url.Values,
	key string,
	defaultValue string,
) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readQueryInt(
	qs url.Values,
	key string,
	defaultValue int,
	v *validator.Validator,
) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func (app *application) readQueryParamToIntPtr(
	qs url.Values,
	key string,
	v *validator.Validator,
) *int {
	s := qs.Get(key)

	if s == "" {
		return nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
	}
	return &i
}

func (app *application) readQueryDate(
	qs url.Values,
	key string,
	v *validator.Validator,
) *time.Time {
	s := qs.Get(key)
	if s == "" {
		return nil
	}
	date, err := time.Parse("2006-01-02", s)
	if err != nil {
		v.AddError(key, "not a valid date format ('2006-01-02')")
		return nil
	}

	return &date
}

func (app *application) failedValidationResponse(
	w http.ResponseWriter,
	r *http.Request,
	errors map[string]string,
) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) readQueryCommaSeperatedString(
	qs url.Values,
	key string,
	defaultValue string,
) []string {
	s := qs.Get(key)

	if s == "" {
		return []string{defaultValue}
	}

	splitValues := strings.Split(s, ",")

	var seen []string
	var values []string
	for _, val := range splitValues {
		trimmedVal := strings.TrimSpace(val)
		normalizedVal := strings.TrimPrefix(trimmedVal, "-")
		if trimmedVal != "" && !slices.Contains(seen, normalizedVal) {
			seen = append(seen, normalizedVal)
			values = append(values, trimmedVal)
		}
	}

	return values
}

type ctxKey string

const (
	slogFields ctxKey = "slog_fields"
)

type ContextHandler struct {
	slog.Handler
}

// Handle adds contextual attributes to the Record before calling the underlying
// handler
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(slogFields).([]slog.Attr); ok {
		for _, v := range attrs {
			r.AddAttrs(v)
		}
	}

	return h.Handler.Handle(ctx, r)
}

// AppendCtx adds an slog attribute to the provided context so that it will be
// included in any Record created with such context
func AppendCtx(parent context.Context, attr slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(slogFields).([]slog.Attr); ok {
		v = append(v, attr)
		return context.WithValue(parent, slogFields, v)
	}

	v := []slog.Attr{}
	v = append(v, attr)
	return context.WithValue(parent, slogFields, v)
}
