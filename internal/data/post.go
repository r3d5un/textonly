package data

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"
)

type BlogPost struct {
	ID         int        `json:"id"`
	Title      string     `json:"title"`
	Lead       string     `json:"lead"`
	Post       string     `json:"post"`
	LastUpdate *time.Time `json:"last_update,omitempty"`
	Created    *time.Time `json:"created,omitempty"`
}

type BlogPostModel struct {
	DB *sql.DB
}

func (m *BlogPostModel) Get(id int) (*BlogPost, error) {
	stmt := `
        SELECT id, title, lead, post, last_update, created
        FROM posts
        WHERE id = $1;`

	slog.Info("querying blogpost", "query", stmt, "id", id)
	row := m.DB.QueryRow(stmt, id)
	blogPost := &BlogPost{}

	err := row.Scan(
		&blogPost.ID,
		&blogPost.Title,
		&blogPost.Lead,
		&blogPost.Post,
		&blogPost.LastUpdate,
		&blogPost.Created,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("no records found", "query", stmt, "id", id)
			return nil, ErrNoRecord
		} else {
			slog.Info("unable to query blogpost", "query", stmt, "id", id)
			return nil, err
		}
	}
	return blogPost, nil
}

func (m *BlogPostModel) GetAll(filters Filters) ([]*BlogPost, Metadata, error) {
	stmt := `
        SELECT id, title, lead, post, last_update, created
        FROM posts
        ORDER BY id DESC;`

	slog.Info("querying blogposts", "query", stmt)
	rows, err := m.DB.Query(stmt)
	if err != nil {
		slog.Error("unable to query blogposts", "query", stmt, "error", err)
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	blogPosts := []*BlogPost{}
	for rows.Next() {
		blogPost := &BlogPost{}
		err = rows.Scan(
			&blogPost.ID,
			&blogPost.Title,
			&blogPost.Lead,
			&blogPost.Post,
			&blogPost.LastUpdate,
			&blogPost.Created,
		)
		if err != nil {
			slog.Error("unable to query blogposts", "query", stmt, "error", err)
			return nil, Metadata{}, err
		}
		blogPosts = append(blogPosts, blogPost)
	}
	if err = rows.Err(); err != nil {
		slog.Error("unable to query blogposts", "query", stmt, "error", err)
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize, filters.OrderBy)

	return blogPosts, metadata, nil
}

func (m *BlogPostModel) LastN(limit int) ([]*BlogPost, error) {
	stmt := `
        SELECT id, title, lead, post, last_update, created
        FROM posts
        ORDER BY id DESC
        LIMIT $1;`

	slog.Info("querying last blogposts", "query", stmt, "limit", limit)
	rows, err := m.DB.Query(stmt, limit)
	if err != nil {
		slog.Info("unable to query last blogposts", "query", stmt, "limit", limit)
		return nil, err
	}
	defer rows.Close()

	blogPosts := []*BlogPost{}
	for rows.Next() {
		blogPost := &BlogPost{}
		err = rows.Scan(
			&blogPost.ID,
			&blogPost.Title,
			&blogPost.Lead,
			&blogPost.Post,
			&blogPost.LastUpdate,
			&blogPost.Created,
		)
		if err != nil {
			slog.Info("unable to query last blogposts", "query", stmt, "limit", limit)
			return nil, err
		}
		blogPosts = append(blogPosts, blogPost)
	}
	if err = rows.Err(); err != nil {
		slog.Info("unable to query last blogposts", "query", stmt, "limit", limit)
		return nil, err
	}

	return blogPosts, nil
}

func (m *BlogPostModel) Insert(bp *BlogPost) (BlogPost, error) {
	query := `INSERT INTO posts (
        title, lead, post
        )
        VALUES ($1, $2, $3)
        RETURNING id, last_update, created;`

	args := []any{
		bp.Title,
		bp.Lead,
		bp.Post,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&bp.ID, &bp.LastUpdate, &bp.Created,
	)
	if err != nil {
		return *bp, err
	}

	return *bp, nil
}

func (m *BlogPostModel) Update(bp *BlogPost) error {
	query := `UPDATE posts
        SET title = $2, lead = $3, post = $4, last_update = NOW(), created = $5
        WHERE id = $1
    `

	args := []any{
		bp.ID,
		bp.Title,
		bp.Lead,
		bp.Post,
		bp.Created,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNoRecord
		default:
			return err
		}
	}

	return nil
}

func (m *BlogPostModel) Delete(id int) (rowsAffected int64, err error) {
	query := "DELETE FROM posts WHERE id = $1;"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return 0, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		return 0, ErrRecordNotFound
	}

	return rowsAffected, nil
}
