package main

import (
	"github.com/russross/blackfriday/v2"
	"html/template"
	"path/filepath"
	"regexp"
	"textonly.islandwind.me/internal/models"
	"time"
)

type templateData struct {
	BlogPost  *models.BlogPost
	BlogPosts []*models.BlogPost
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func humanDate(t time.Time) string {
	return t.Format("2006-01-02 15:04")
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
