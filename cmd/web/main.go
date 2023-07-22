package main

import (
	"database/sql"
	"flag"
	_ "github.com/jackc/pgx/v5/stdlib"
	"html/template"
	"log"
	"net/http"
	"os"
	"textonly.islandwind.me/internal/models"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	blogPosts     *models.BlogPostModel
	templateCache map[string]*template.Template
	feedCache     map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String(
		"dsn",
		"user=postgres "+
			"password=postgres "+
			"host=localhost "+
			"port=5432 "+
			"dbname=blog "+
			"sslmode=disable ",
		"PostgreSQL data source DSN",
	)
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime)

	db, err := openDB(*dsn)
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
		blogPosts:     &models.BlogPostModel{DB: db},
		templateCache: templateCache,
		feedCache:     feedCache,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
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
