package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"textonly.islandwind.me/cmd/web/config"
	"textonly.islandwind.me/internal/data"
	"textonly.islandwind.me/internal/vcs"
)

var (
	version = vcs.Version()
)

type application struct {
	logger        *slog.Logger
	models        data.Models
	templateCache map[string]*template.Template
	config        *config.Config
}

func main() {
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("textonly.islandwind.me %s\n", version)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("loading configuration")
	config, err := config.New()
	if err != nil {
		slog.Error("unable to load configuration", "error", err)
	}
	slog.Info("configuration loaded", "environment", config.App.ENV)

	logger.Info("opening database connection pool...")
	db, err := openDB(config.Database.DSN)
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
		models:        data.NewModels(db),
		templateCache: templateCache,
		config:        config,
	}

	err = app.serve(app.config.App.URL)
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
