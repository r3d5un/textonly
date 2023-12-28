package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	BlogPosts BlogPostModel
	Socials   SocialModel
	Users     UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		BlogPosts: BlogPostModel{DB: db},
		Socials:   SocialModel{DB: db},
		Users:     UserModel{DB: db},
	}
}
