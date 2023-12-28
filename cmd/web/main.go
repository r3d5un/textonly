package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"textonly.islandwind.me/internal/data"
	"textonly.islandwind.me/internal/vcs"
)

var (
	version = vcs.Version()
)

type application struct {
	logger        *slog.Logger
	blogPosts     *models.BlogPostModel
	sosials       *models.SocialModel
	user          *models.UserModel
	templateCache map[string]*template.Template
}

func main() {
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("textonly.islandwind.me %s\n", version)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	addr := GetURL()
	dsn := GetDBDSN()

	logger.Info("opening database connection pool...")
	db, err := openDB(dsn)
	if err != nil {
		logger.Error("unable to open database connection pool", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	logger.Info("caching templates...")
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error("an error occurred while caching templates", "error", err)
		os.Exit(1)
	}
	logger.Info("templates successfully cached")

	app := &application{
		logger:        logger,
		blogPosts:     &models.BlogPostModel{DB: db},
		sosials:       &models.SocialModel{DB: db},
		user:          &models.UserModel{DB: db},
		templateCache: templateCache,
	}

	err = app.serve(addr)
	if err != nil {
		logger.Error("an error occurred", "error", err)
		os.Exit(1)
	}
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
