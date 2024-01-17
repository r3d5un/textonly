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
	Timeout *time.Duration
	DB      *sql.DB
	Logger  *slog.Logger
}

func (m *BlogPostModel) Get(ctx context.Context, id int) (*BlogPost, error) {
	if id < 1 {
		m.Logger.InfoContext(ctx, "invalid id", "id", id)
		return nil, ErrRecordNotFound
	}
	stmt := `
        SELECT id, title, lead, post, last_update, created
        FROM posts
        WHERE id = $1;`

	qCtx, cancel := context.WithTimeout(ctx, *m.Timeout)
	defer cancel()

	m.Logger.InfoContext(qCtx, "querying blogpost", "query", stmt, "id", id)
	row := m.DB.QueryRowContext(ctx, stmt, id)
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
			m.Logger.InfoContext(ctx, "no records found", "query", stmt, "id", id)
			return nil, ErrNoRecord
		} else {
			m.Logger.InfoContext(ctx, "unable to query blogpost", "query", stmt, "id", id)
			return nil, err
		}
	}
	m.Logger.InfoContext(ctx, "data retrieved")

	return blogPost, nil
}

func (m *BlogPostModel) GetAll(
	ctx context.Context,
	filters Filters,
) ([]*BlogPost, Metadata, error) {
	stmt := `
        SELECT COUNT(*) OVER(), id, title, lead, post, last_update, created
        FROM posts
        WHERE
            ($1::int IS NULL OR id = $1)
            AND ($2 = '' OR title LIKE ('%' || $2 || '%'))
            AND ($3 = '' OR lead LIKE ('%' || $3 || '%'))
            AND ($4 = '' OR post LIKE ('%' || $4 || '%'))
            AND ($5::timestamp IS NULL OR created >= $5)
            AND ($6::timestamp IS NULL OR created <= $6)
            AND ($7::timestamp IS NULL OR last_update >= $7)
            AND ($8::timestamp IS NULL OR last_update <= $8)
        ` + CreateOrderByClause(filters.OrderBy) + `
        LIMIT $9 OFFSET $10;`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	m.Logger.InfoContext(ctx, "querying blogposts", "query", stmt, "filters", filters)
	rows, err := m.DB.QueryContext(
		ctx,
		stmt,
		filters.ID,
		filters.Title,
		filters.Lead,
		filters.Post,
		filters.CreatedFrom,
		filters.CreatedTo,
		filters.LastUpdatedFrom,
		filters.LastUpdatedTo,
		filters.limit(),
		filters.offset(),
	)
	if err != nil {
		m.Logger.ErrorContext(ctx, "unable to query blogposts", "query", stmt, "error", err)
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	blogPosts := []*BlogPost{}
	for rows.Next() {
		blogPost := &BlogPost{}
		err = rows.Scan(
			&totalRecords,
			&blogPost.ID,
			&blogPost.Title,
			&blogPost.Lead,
			&blogPost.Post,
			&blogPost.LastUpdate,
			&blogPost.Created,
		)
		if err != nil {
			m.Logger.ErrorContext(ctx, "unable to query blogposts", "query", stmt, "error", err)
			return nil, Metadata{}, err
		}
		blogPosts = append(blogPosts, blogPost)
	}
	if err = rows.Err(); err != nil {
		m.Logger.ErrorContext(ctx, "unable to query blogposts", "query", stmt, "error", err)
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize, filters.OrderBy)
	m.Logger.InfoContext(ctx, "data retrieved", "metadata", metadata)

	return blogPosts, metadata, nil
}

func (m *BlogPostModel) LastN(ctx context.Context, limit int) ([]*BlogPost, error) {
	stmt := `
        SELECT id, title, lead, post, last_update, created
        FROM posts
        ORDER BY id DESC
        LIMIT $1;`

	m.Logger.InfoContext(ctx, "querying last blogposts", "query", stmt, "limit", limit)
	rows, err := m.DB.Query(stmt, limit)
	if err != nil {
		m.Logger.Info("unable to query last blogposts", "query", stmt, "limit", limit)
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
			m.Logger.InfoContext(
				ctx,
				"unable to query last blogposts",
				"query",
				stmt,
				"limit",
				limit,
			)
			return nil, err
		}
		blogPosts = append(blogPosts, blogPost)
	}
	if err = rows.Err(); err != nil {
		m.Logger.InfoContext(ctx, "unable to query last blogposts", "query", stmt, "limit", limit)
		return nil, err
	}
	m.Logger.InfoContext(ctx, "data retrieved")

	return blogPosts, nil
}

func (m *BlogPostModel) Insert(ctx context.Context, bp *BlogPost) (BlogPost, error) {
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

	rCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	m.Logger.InfoContext(rCtx, "inserting blogpost", "query", query, "args", args)
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&bp.ID, &bp.LastUpdate, &bp.Created,
	)
	if err != nil {
		m.Logger.ErrorContext(
			ctx,
			"unable to insert blogpost",
			"query",
			query,
			"args",
			args,
			"error",
			err,
		)
		return *bp, err
	}
	m.Logger.InfoContext(ctx, "blogpost inserted", "id", bp.ID)

	return *bp, nil
}

func (m *BlogPostModel) Update(ctx context.Context, bp *BlogPost) (rowsAffected int64, err error) {
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

	rCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	m.Logger.InfoContext(ctx, "updating blogpost", "query", query, "args", args)
	result, err := m.DB.ExecContext(rCtx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			m.Logger.InfoContext(ctx, "no records found", "query", query, "args", args)
			return 0, ErrNoRecord
		default:
			m.Logger.ErrorContext(
				ctx,
				"unable to update blogpost",
				"query",
				query,
				"args",
				args,
				"error",
				err,
			)
			return 0, err
		}
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		m.Logger.ErrorContext(
			ctx,
			"unable to update blogpost",
			"query",
			query,
			"args",
			args,
			"error",
			err,
		)
		return 0, err
	}
	if rowsAffected == 0 {
		m.Logger.InfoContext(ctx, "no records found", "query", query, "args", args)
		return 0, ErrRecordNotFound
	}
	m.Logger.InfoContext(ctx, "blogpost updated", "id", bp.ID)

	return rowsAffected, nil
}

func (m *BlogPostModel) Delete(ctx context.Context, id int) (rowsAffected int64, err error) {
	query := "DELETE FROM posts WHERE id = $1;"

	rCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	m.Logger.InfoContext(rCtx, "deleting blogpost", "query", query, "id", id)
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		m.Logger.ErrorContext(
			ctx,
			"unable to delete blogpost",
			"query",
			query,
			"id",
			id,
			"error",
			err,
		)
		return 0, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		m.Logger.ErrorContext(
			ctx,
			"unable to delete blogpost",
			"query",
			query,
			"id",
			id,
			"error",
			err,
		)
		return 0, err
	}
	if rowsAffected == 0 {
		m.Logger.InfoContext(ctx, "no records found", "query", query, "id", id)
		return 0, ErrRecordNotFound
	}
	m.Logger.InfoContext(ctx, "blogpost deleted", "id", id)

	return rowsAffected, nil
}
