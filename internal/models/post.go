package models

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

type BlogPost struct {
	ID         int
	Title      string
	Lead       string
	Post       string
	LastUpdate time.Time
	Created    time.Time
}

type BlogPostModel struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func (m *BlogPostModel) Get(id int) (*BlogPost, error) {
	stmt := `
        SELECT id, title, lead, post, last_update, created
        FROM posts
        WHERE id = $1;`
	m.InfoLog.Print("query statement: ", stmt)

	row := m.DB.QueryRow(stmt, id)
	blogPost := &BlogPost{}

	err := row.Scan(&blogPost.ID, &blogPost.Title, &blogPost.Lead, &blogPost.Post, &blogPost.LastUpdate, &blogPost.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return blogPost, nil
}

func (m *BlogPostModel) GetAll() ([]*BlogPost, error) {
	stmt := `
        SELECT id, title, lead, post, last_update, created
        FROM posts
        ORDER BY id DESC;`
	m.InfoLog.Print("query statement: ", stmt)

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blogPosts := []*BlogPost{}
	for rows.Next() {
		blogPost := &BlogPost{}
		err = rows.Scan(&blogPost.ID, &blogPost.Title, &blogPost.Lead, &blogPost.Post, &blogPost.LastUpdate, &blogPost.Created)
		if err != nil {
			return nil, err
		}
		blogPosts = append(blogPosts, blogPost)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return blogPosts, nil
}

func (m *BlogPostModel) LastN(limit int) ([]*BlogPost, error) {
	stmt := `
        SELECT id, title, lead, post, last_update, created
        FROM posts
        ORDER BY id DESC
        LIMIT $1;`
	m.InfoLog.Print("query statement: ", stmt)

	rows, err := m.DB.Query(stmt, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blogPosts := []*BlogPost{}
	for rows.Next() {
		blogPost := &BlogPost{}
		err = rows.Scan(&blogPost.ID, &blogPost.Title, &blogPost.Lead, &blogPost.Post, &blogPost.LastUpdate, &blogPost.Created)
		if err != nil {
			return nil, err
		}
		blogPosts = append(blogPosts, blogPost)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return blogPosts, nil
}
