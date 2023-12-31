package data

import (
	"database/sql"
	"errors"
	"log/slog"
)

type User struct {
	ID      int
	Name    string
	Summary string
	Content string
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
