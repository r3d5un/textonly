package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"textonly.islandwind.me/internal/models"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	blogPosts     *models.BlogPostModel
	sosials       *models.SocialModel
	user          *models.UserModel
	templateCache map[string]*template.Template
	feedCache     map[string]*template.Template
}

func main() {
	addr := GetURL()
	dsn := GetDBDSN()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime)

	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}
	feedCache, err := newFeedTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		blogPosts:     &models.BlogPostModel{DB: db, InfoLog: infoLog, ErrorLog: errorLog},
		sosials:       &models.SocialModel{DB: db, InfoLog: infoLog, ErrorLog: errorLog},
		user:          &models.UserModel{DB: db, InfoLog: infoLog, ErrorLog: errorLog},
		templateCache: templateCache,
		feedCache:     feedCache,
	}

	srv := &http.Server{
		Addr:     addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
