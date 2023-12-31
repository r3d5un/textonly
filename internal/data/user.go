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
	DB *sql.DB
}

func (m *UserModel) Get(id int) (*User, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	stmt := `
        SELECT user_id, name, summary, content
        FROM users
        WHERE user_id = $1
        LIMIT 1;`

	slog.Info("querying database for user", "query", stmt, "id", id)
	row := m.DB.QueryRow(stmt, id)
	user := &User{}

	err := row.Scan(&user.ID, &user.Name, &user.Summary, &user.Content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("no records found", "error", err)
			return nil, ErrNoRecord
		} else {
			slog.Error("unable to query user", "error", err)
			return nil, err
		}
	}
	return user, nil
}

func (m *UserModel) Update(u *User) (rowsAffected int64, err error) {
	query := `UPDATE users
    SET name = $2, summary = $3, content = $4
    WHERE user_id = $1;`

	args := []any{
		u.ID,
		u.Name,
		u.Summary,
		u.Content,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrNoRecord
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
