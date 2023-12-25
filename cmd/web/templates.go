package main

import (
	"html/template"
	"io/fs"
	"log/slog"
	"path/filepath"
	"regexp"
	"time"

	"github.com/russross/blackfriday/v2"
	"textonly.islandwind.me/internal/models"
	"textonly.islandwind.me/ui"
)

type templateData struct {
	BlogPost  *models.BlogPost
	BlogPosts []*models.BlogPost
	Socials   []*models.Social
	User      *models.User
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		slog.Error("an error occurred while walking template directory", "error", err)
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}
		slog.Info("parsed template", "template", name)

		cache[name] = ts
	}

	feedTemplate, err := template.ParseFS(ui.Files, "xml/feed.tmpl")
	if err != nil {
		slog.Error("an error occurred when collecting RSS feed template", "error", err)
		return nil, err
	}
	slog.Info("parsed RSS feed template")

	cache["feed"] = feedTemplate

	return cache, nil
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04")
}

func removeMarkdownTitle(input string) string {
	re := regexp.MustCompile(`(?m)^#([^#].*)`)
	return re.ReplaceAllString(input, "")
}

func markdownToHTML(input string) template.HTML {
	return template.HTML(blackfriday.Run([]byte(removeMarkdownTitle(input))))
}

var functions = template.FuncMap{
	"humanDate":      humanDate,
	"markdownToHTML": markdownToHTML,
}
