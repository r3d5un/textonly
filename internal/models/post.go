package models

import (
	"database/sql"
	"time"
)

type BlogPost struct {
	ID         int
	Title      string
	Post       string
	LastUpdate time.Time
	Created    time.Time
}

type BlogPostModel struct {
	DB *sql.DB
}

func (m *BlogPostModel) Get(id int) (*BlogPost, error) {
	return nil, nil
}

func (m *BlogPostModel) LastN(limit int) ([]*BlogPost, error) {
	return nil, nil
}
