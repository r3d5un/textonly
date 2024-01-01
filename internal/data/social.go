package data

import (
	"database/sql"
	"errors"
	"log/slog"
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
