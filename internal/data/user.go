package data

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"
)

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Content string `json:"content"`
}

type UserModel struct {
	Timeout *time.Duration
	DB      *sql.DB
	Logger  *slog.Logger
}

func (m *UserModel) Get(ctx context.Context, id int) (*User, error) {
	if id < 1 {
		m.Logger.InfoContext(ctx, "invalid id", "id", id)
		return nil, ErrRecordNotFound
	}

	stmt := `
        SELECT user_id, name, summary, content
        FROM users
        WHERE user_id = $1
        LIMIT 1;`

	rCtx, cancel := context.WithTimeout(ctx, *m.Timeout)
	defer cancel()

	m.Logger.Info("querying database for user", "query", stmt, "id", id)
	row := m.DB.QueryRowContext(rCtx, stmt, id)
	user := &User{}

	err := row.Scan(&user.ID, &user.Name, &user.Summary, &user.Content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			m.Logger.InfoContext(ctx, "no records found", "error", err)
			return nil, ErrNoRecord
		} else {
			m.Logger.ErrorContext(ctx, "unable to query user", "error", err)
			return nil, err
		}
	}
	m.Logger.InfoContext(ctx, "data retrieved")

	return user, nil
}

func (m *UserModel) Update(ctx context.Context, u *User) (rowsAffected int64, err error) {
	query := `UPDATE users
    SET name = $2, summary = $3, content = $4
    WHERE user_id = $1;`

	args := []any{
		u.ID,
		u.Name,
		u.Summary,
		u.Content,
	}

	rCtx, cancel := context.WithTimeout(ctx, *m.Timeout)
	defer cancel()

	m.Logger.InfoContext(ctx, "updating user", "query", query, "args", args)
	result, err := m.DB.ExecContext(rCtx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			m.Logger.InfoContext(ctx, "no records found", "error", err)
			return 0, ErrNoRecord
		default:
			m.Logger.ErrorContext(ctx, "unable to update user", "error", err)
			return 0, err
		}
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		m.Logger.ErrorContext(ctx, "unable to update user", "error", err)
		return 0, err
	}
	if rowsAffected == 0 {
		m.Logger.InfoContext(ctx, "no records found", "error", err)
		return 0, ErrRecordNotFound
	}
	m.Logger.InfoContext(ctx, "user updated", "rows_affected", rowsAffected)

	return rowsAffected, nil
}
