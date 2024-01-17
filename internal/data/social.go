package data

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"
)

type Social struct {
	ID             int    `json:"id"`
	UserID         int    `json:"user_id"`
	SocialPlatform string `json:"social_platform"`
	Link           string `json:"link"`
}

type SocialModel struct {
	Timeout *time.Duration
	DB      *sql.DB
	Logger  *slog.Logger
}

func (m *SocialModel) Get(ctx context.Context, id int) (*Social, error) {
	if id < 1 {
		m.Logger.InfoContext(ctx, "invalid id", "id", id)
		return nil, ErrRecordNotFound
	}

	stmt := `SELECT id, user_id, social_platform, link
        FROM socials
        WHERE id = $1;`

	rCtx, cancel := context.WithTimeout(ctx, *m.Timeout)
	defer cancel()

	m.Logger.InfoContext(ctx, "querying social data", "query", stmt, "id", id)
	row := m.DB.QueryRowContext(rCtx, stmt, id)
	s := &Social{}

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.SocialPlatform,
		&s.Link,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			m.Logger.InfoContext(ctx, "no records found", "query", stmt, "id", id)
			return nil, ErrRecordNotFound
		} else {
			m.Logger.InfoContext(ctx, "unable to query social data", "query", stmt, "id", id)
			return nil, err
		}
	}
	m.Logger.InfoContext(ctx, "data retrieved")

	return s, nil
}

func (m *SocialModel) GetByUserID(ctx context.Context, id int) ([]*Social, error) {
	stmt := `
        SELECT id, user_id, social_platform, link
        FROM socials
        WHERE user_id = $1;`

	rCtx, cancel := context.WithTimeout(ctx, *m.Timeout)
	defer cancel()

	m.Logger.InfoContext(ctx, "querying socials", "query", stmt, "id", id)
	rows, err := m.DB.QueryContext(rCtx, stmt, id)
	if err != nil {
		m.Logger.ErrorContext(ctx, "unable to query socials", "query", stmt, "id", id, "error", err)
		return nil, err
	}
	defer rows.Close()

	socials := []*Social{}
	for rows.Next() {
		social := &Social{}
		err = rows.Scan(
			&social.ID,
			&social.UserID,
			&social.SocialPlatform,
			&social.Link,
		)
		if err != nil {
			m.Logger.InfoContext(ctx, "no records found", "query", stmt, "id", id)
			return nil, err
		}
		socials = append(socials, social)
	}
	if err = rows.Err(); err != nil {
		m.Logger.InfoContext(ctx, "unable to query social data", "query", stmt, "id", id)
		return nil, err
	}
	m.Logger.InfoContext(ctx, "data retrieved")

	return socials, nil
}

func (m *SocialModel) GetAll(
	ctx context.Context,
	filters Filters,
) ([]*Social, Metadata, error) {
	stmt := `SELECT COUNT(*) OVER(), id, user_id, social_platform, link
    FROM socials
    WHERE
        ($1::int IS NULL OR id = $1)
        AND ($2::int IS NULL OR user_id = $2)
        AND ($3 = '' OR social_platform LIKE ('%' || $3 || '%'))
    ` + CreateOrderByClause(filters.OrderBy) + `
    LIMIT $4 OFFSET $5;`

	rCtx, cancel := context.WithTimeout(ctx, *m.Timeout)
	defer cancel()

	m.Logger.InfoContext(rCtx, "querying social data", "query", stmt, "filters", filters)
	rows, err := m.DB.QueryContext(
		ctx,
		stmt,
		filters.ID,
		filters.UserID,
		filters.SocialPlatform,
		filters.limit(),
		filters.offset(),
	)
	if err != nil {
		m.Logger.ErrorContext(ctx, "unable to query blogposts", "query", stmt, "error", err)
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	socials := []*Social{}

	for rows.Next() {
		social := &Social{}
		err = rows.Scan(
			&totalRecords,
			&social.ID,
			&social.UserID,
			&social.SocialPlatform,
			&social.Link,
		)
		if err != nil {
			m.Logger.ErrorContext(ctx, "unable to query socials", "query", stmt, "error", err)
			return nil, Metadata{}, err
		}
		socials = append(socials, social)
	}

	if err = rows.Err(); err != nil {
		m.Logger.InfoContext(ctx, "unable to query social data", "query", stmt, "filters", filters)
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize, filters.OrderBy)
	m.Logger.InfoContext(ctx, "data retrieved", "metadata", metadata)

	return socials, metadata, nil
}

func (m *SocialModel) Insert(ctx context.Context, s *Social) (Social, error) {
	query := `INSERT INTO socials (
        user_id, social_platform, link
    )
    VALUES ($1, $2, $3)
    RETURNING id, user_id, social_platform, link;`

	args := []any{s.UserID, s.SocialPlatform, s.Link}

	rCtx, cancel := context.WithTimeout(ctx, *m.Timeout)
	defer cancel()

	m.Logger.InfoContext(rCtx, "inserting social data", "query", query, "args", args)
	err := m.DB.QueryRowContext(rCtx, query, args...).Scan(
		&s.UserID, &s.SocialPlatform, &s.Link,
	)
	if err != nil {
		return *s, err
	}
	m.Logger.InfoContext(ctx, "blogpost inserted", "social_data", *s)

	return *s, err
}

func (m *SocialModel) Update(ctx context.Context, s *Social) (rowsAffected int64, err error) {
	query := `UPDATE socials
    SET user_id = $2, social_platform = $3, link = $4
    WHERE id = $1;`

	args := []any{
		s.ID,
		s.UserID,
		s.SocialPlatform,
		s.Link,
	}

	rCtx, cancel := context.WithTimeout(ctx, *m.Timeout)
	defer cancel()

	m.Logger.InfoContext(rCtx, "updating social data", "query", query, "args", args)
	result, err := m.DB.ExecContext(rCtx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			m.Logger.InfoContext(ctx, "no records found", "query", query, "args", args)
			return 0, ErrRecordNotFound
		default:
			m.Logger.InfoContext(ctx, "unable to query social data", "query", query, "args", args)
			return 0, err
		}
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		m.Logger.InfoContext(ctx, "unable to query social data", "query", query, "args", args)
		return 0, err
	}
	if rowsAffected == 0 {
		m.Logger.InfoContext(ctx, "no records found", "query", query, "args", args)
		return 0, ErrRecordNotFound
	}
	m.Logger.InfoContext(ctx, "social data updated", "query", query, "args", args)

	return rowsAffected, nil
}

func (m *SocialModel) Delete(ctx context.Context, id int) (rowsAffected int64, err error) {
	query := "DELETE FROM socials WHERE id = $1;"

	rCtx, cancel := context.WithTimeout(ctx, *m.Timeout)
	defer cancel()

	m.Logger.InfoContext(rCtx, "deleting social data", "query", query, "id", id)
	result, err := m.DB.ExecContext(rCtx, query, id)
	if err != nil {
		m.Logger.ErrorContext(ctx, "unable to delete social data", "id", id, "error", err)
		return 0, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		m.Logger.ErrorContext(ctx, "unable to delete social data", "id", id, "error", err)
		return 0, err
	}
	if rowsAffected == 0 {
		m.Logger.InfoContext(ctx, "no records found", "id", id)
		return 0, ErrRecordNotFound
	}
	m.Logger.InfoContext(ctx, "social data deleted", "id", id)

	return rowsAffected, nil
}
