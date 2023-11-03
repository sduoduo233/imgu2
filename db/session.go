package db

import (
	"fmt"
	"time"
)

func SessionCreate(token string, userId int, exipreAt int64) error {
	_, err := DB.Exec("INSERT INTO sessions(token, user, expire_at) VALUES(?, ?, ?)", token, userId, exipreAt)
	return err
}

// returns sql.ErrNoRows if token is invalid
func SessionFind(token string) (int, error) {
	row := DB.QueryRow("SELECT (user) FROM sessions WHERE token = ? AND expire_at > ?", token, time.Now().Unix())

	var userId int
	err := row.Scan(&userId)
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}
	return userId, nil
}

func SessionDelete(token string) error {
	_, err := DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}
