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
	DB *sql.DB
}

func (m *SocialModel) Get(id int) (*Social, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	stmt := `SELECT id, user_id, social_platform, link
        FROM socials
        WHERE id = $1;`

	slog.Info("querying social data", "query", stmt, "id", id)
	row := m.DB.QueryRow(stmt, id)
	s := &Social{}

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.SocialPlatform,
		&s.Link,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("no records found", "query", stmt, "id", id)
			return nil, ErrRecordNotFound
		} else {
			slog.Info("unable to query social", "query", stmt, "id", id)
			return nil, err
		}
	}

	return s, nil
}

func (m *SocialModel) GetByUserID(id int) ([]*Social, error) {
	stmt := `
        SELECT id, user_id, social_platform, link
        FROM socials
        WHERE user_id = $1;`

	slog.Info("querying socials", "query", stmt, "id", id)
	rows, err := m.DB.Query(stmt, id)
	if err != nil {
		slog.Error("unable to query socials", "query", stmt, "id", id, "error", err)
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
			slog.Error("unable to query socials", "query", stmt, "id", id, "error", err)
			return nil, err
		}
		socials = append(socials, social)
	}
	if err = rows.Err(); err != nil {
		slog.Error("unable to query socials", "query", stmt, "id", id, "error", err)
		return nil, err
	}

	return socials, nil
}

func (m *SocialModel) GetAll(filters Filters) ([]*Social, Metadata, error) {
	stmt := `SELECT COUNT(*) OVER(), id, user_id, social_platform, link
    FROM socials
    WHERE
        ($1::int IS NULL OR id = $1)
        AND ($2::int IS NULL OR user_id = $2)
        AND ($3 = '' OR social_platform LIKE ('%' || $3 || '%'))
    ORDER BY id DESC
    LIMIT $4 OFFSET $5;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	slog.Info("querying social data", "query", stmt, "filters", filters)
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
			slog.Error("unable to query socials", "query", stmt, "error", err)
			return nil, Metadata{}, err
		}
		socials = append(socials, social)
	}

	if err = rows.Err(); err != nil {
		slog.Error("unable to query socials", "query", stmt, "error", err)
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize, filters.OrderBy)

	return socials, metadata, nil
}

func (m *SocialModel) Insert(s *Social) (Social, error) {
	query := `INSERT INTO socials (
        user_id, social_platform, link
    )
    VALUES ($1, $2, $3)
    RETURNING id, user_id, social_platform, link;`

	args := []any{s.UserID, s.SocialPlatform, s.Link}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&s.UserID, &s.SocialPlatform, &s.Link,
	)
	if err != nil {
		return *s, err
	}

	return *s, err
}

func (m *SocialModel) Update(s *Social) (rowsAffected int64, err error) {
	query := `UPDATE socials
    SET user_id = $2, social_platform = $3, link = $4
    WHERE id = $1;`

	args := []any{
		s.ID,
		s.UserID,
		s.SocialPlatform,
		s.Link,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrRecordNotFound
		default:
			return 0, err
		}
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

func (m *SocialModel) Delete(id int) (rowsAffected int64, err error) {
	query := "DELETE FROM socials WHERE id = $1;"

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
