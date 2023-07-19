package models

import (
	"database/sql"
	"errors"
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
	stmt := `
        SELECT id, title, post, last_update, created
        FROM posts
        WHERE id = $1;
    `
	row := m.DB.QueryRow(stmt, id)
	blogPost := &BlogPost{}

	err := row.Scan(&blogPost.ID, &blogPost.Title, &blogPost.Post, &blogPost.LastUpdate, &blogPost.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return blogPost, nil
}

func (m *BlogPostModel) LastN(limit int) ([]*BlogPost, error) {
	return nil, nil
}
