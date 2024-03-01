package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"textonly.islandwind.me/internal/utils"
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
}

func (m *UserModel) Get(ctx context.Context, id int) (*User, error) {
	logger := utils.LoggerFromContext(ctx)

	if id < 1 {
		logger.InfoContext(ctx, "invalid id", "id", id)
		return nil, ErrRecordNotFound
	}

	stmt := `
        SELECT user_id, name, summary, content
        FROM users
        WHERE user_id = $1
        LIMIT 1;`

	rCtx, cancel := context.WithTimeout(ctx, *m.Timeout)
	defer cancel()

	logger.Info("querying database for user", "query", stmt, "id", id)
	row := m.DB.QueryRowContext(rCtx, stmt, id)
	user := &User{}

	err := row.Scan(&user.ID, &user.Name, &user.Summary, &user.Content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.InfoContext(ctx, "no records found", "error", err)
			return nil, ErrNoRecord
		} else {
			logger.ErrorContext(ctx, "unable to query user", "error", err)
			return nil, err
		}
	}
	logger.InfoContext(ctx, "data retrieved")

	return user, nil
}

func (m *UserModel) Update(ctx context.Context, u *User) (rowsAffected int64, err error) {
	logger := utils.LoggerFromContext(ctx)

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

	logger.InfoContext(ctx, "updating user", "query", query, "args", args)
	result, err := m.DB.ExecContext(rCtx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			logger.InfoContext(ctx, "no records found", "error", err)
			return 0, ErrNoRecord
		default:
			logger.ErrorContext(ctx, "unable to update user", "error", err)
			return 0, err
		}
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		logger.ErrorContext(ctx, "unable to update user", "error", err)
		return 0, err
	}
	if rowsAffected == 0 {
		logger.InfoContext(ctx, "no records found", "error", err)
		return 0, ErrRecordNotFound
	}
	logger.InfoContext(ctx, "user updated", "rows_affected", rowsAffected)

	return rowsAffected, nil
}
