package data

import (
	"database/sql"
	"errors"
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

func NewModels(db *sql.DB, timeout *time.Duration) Models {
	return Models{
		BlogPosts: BlogPostModel{DB: db, Timeout: timeout},
		Socials:   SocialModel{DB: db, Timeout: timeout},
		Users:     UserModel{DB: db, Timeout: timeout},
	}
}
