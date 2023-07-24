package models

import (
	"database/sql"
	"errors"
	"log"
)

type User struct {
	ID      int
	Name    string
	Summary string
	Content string
}

type UserModel struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

func (m *UserModel) Get(id int) (*User, error) {
	stmt := `
        SELECT user_id, name, summary, content
        FROM users
        WHERE user_id = $1
        LIMIT 1;`
	m.InfoLog.Print("query statement: ", stmt)

	row := m.DB.QueryRow(stmt, id)
	user := &User{}

	err := row.Scan(&user.ID, &user.Name, &user.Summary, &user.Content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return user, nil
}
