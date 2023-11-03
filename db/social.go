package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type SocialLogin struct {
	Id        int
	Type      string
	UserId    int
	AccountId string
}

const (
	SocialLoginGoogle = "google"
	SocialLoginGithub = "github"
)

func SocialLoginFind(loginType string, userId int) (*SocialLogin, error) {
	if loginType != SocialLoginGithub && loginType != SocialLoginGoogle {
		panic("invalid login type: " + loginType)
	}

	row := DB.QueryRow("SELECT id, type, user, account_id FROM social_logins WHERE type = ? AND user = ? LIMIT 1", loginType, userId)

	var s SocialLogin
	err := row.Scan(&s.Id, &s.Type, &s.UserId, &s.AccountId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("db: %w", err)
	}

	return &s, nil
}

// find a record by social login account
func SocialLoginFindByAccount(loginType string, accountId string) (*SocialLogin, error) {
	if loginType != SocialLoginGithub && loginType != SocialLoginGoogle {
		panic("invalid login type: " + loginType)
	}

	row := DB.QueryRow("SELECT id, type, user, account_id FROM social_logins WHERE type = ? AND account_id = ? LIMIT 1", loginType, accountId)

	var s SocialLogin
	err := row.Scan(&s.Id, &s.Type, &s.UserId, &s.AccountId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("db: %w", err)
	}

	return &s, nil
}

func SocialLoginCreate(loginType string, userId int, accountId string) error {
	if loginType != SocialLoginGithub && loginType != SocialLoginGoogle {
		panic("invalid login type: " + loginType)
	}

	_, err := DB.Exec("INSERT INTO social_logins(type, user, account_id) VALUES(?, ?, ?)", loginType, userId, accountId)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

func SocialLoginRemove(loginType string, userId int) error {
	if loginType != SocialLoginGithub && loginType != SocialLoginGoogle {
		panic("invalid login type: " + loginType)
	}

	_, err := DB.Exec("DELETE FROM social_logins WHERE type = ? AND user = ?", loginType, userId)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}
