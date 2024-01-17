package data

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	BlogPosts BlogPostModel
	Socials   SocialModel
	Users     UserModel
}

func NewModels(db *sql.DB, logger *slog.Logger, timeout *time.Duration) Models {
	return Models{
		BlogPosts: BlogPostModel{DB: db, Logger: logger, Timeout: timeout},
		Socials:   SocialModel{DB: db},
		Users:     UserModel{DB: db},
	}
}
