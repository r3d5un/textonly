package models

import (
	"database/sql"
)

type Social struct {
	ID             int
	UserID         int
	SocialPlatform string
	Link           string
}

type SocialModel struct {
	DB *sql.DB
}

func (m *SocialModel) GetByUserID(id int) ([]*Social, error) {
	stmt := `
        SELECT id, user_id, social_platform, link
        FROM socials
        WHERE user_id = $1
    `
	rows, err := m.DB.Query(stmt, id)
	if err != nil {
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
			return nil, err
		}
		socials = append(socials, social)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return socials, nil
}
