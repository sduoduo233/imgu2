package services

import (
	"database/sql"
	"errors"
	"imgu2/db"
	"time"
)

type session struct{}

var Session = session{}

func (session) Create(userId int) (string, error) {
	token := RandomHexString(8)
	expireAt := time.Now().Add(time.Hour * 24 * 30).Unix()

	err := db.SessionCreate(token, userId, expireAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

// returns nil if token is invalid
func (session) Find(token string) (*db.User, error) {
	userId, err := db.SessionFind(token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	user, err := db.UserFindById(userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (session) Delete(token string) error {
	return db.SessionDelete(token)
}
