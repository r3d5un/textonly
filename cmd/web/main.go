package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
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

// @title			Textonly API
// @version		1.0
// @description	Textonly API
func main() {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	instanceLogger := logger.With(
		slog.Group(
			"application_instance",
			slog.String("version", version),
			slog.String("instance_id", uuid.New().String()),
		),
	)

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("textonly.islandwind.me %s\n", version)
	}

	slog.Info("loading configuration")
	config, err := config.New()
	if err != nil {
		slog.Error("unable to load configuration", "error", err)
		os.Exit(1)
	}
	slog.Info("configuration loaded", "configuration", config)

	logger.Info("opening database connection pool...")
	db, err := openDB(config.Database.DSN)
	if err != nil {
		logger.Error("unable to open database connection pool", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	queryTimeout := time.Duration(config.Database.Timeout) * time.Second
	logger.Info("database connection pool established")

	logger.Info("caching templates...")
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error("an error occurred while caching templates", "error", err)
		os.Exit(1)
	}
	logger.Info("templates successfully cached")

	app := &application{
		logger:        instanceLogger,
		models:        data.NewModels(db, logger, &queryTimeout),
		templateCache: templateCache,
		config:        config,
	}

	err = app.serve(app.config.Server.URL)
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
