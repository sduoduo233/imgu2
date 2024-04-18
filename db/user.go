package db

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	Id            int
	Username      string
	Email         string
	Password      string
	EmailVerified bool
	Role          int
	GroupId       int

	// GroupExpireTime represents the time when the user
	// will no longer be in the user group. Users are
	// automatically changed to the default group
	// by a cron task.
	GroupExpireTime sql.NullTime
}

func UserFindByEmail(email string) (*User, error) {
	var u User
	rows, err := DB.Query("SELECT id, username, email, password, email_verified, role, user_group, user_group_expire FROM users WHERE email = ? LIMIT 1", email)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var groupExpire int

	err = rows.Scan(&u.Id, &u.Username, &u.Email, &u.Password, &u.EmailVerified, &u.Role, &u.GroupId, &groupExpire)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	if groupExpire > 0 {
		u.GroupExpireTime.Valid = true
		u.GroupExpireTime.Time = time.Unix(int64(groupExpire), 0)
	}

	return &u, nil
}

func UserFindById(id int) (*User, error) {
	var u User
	rows, err := DB.Query("SELECT id, username, email, password, email_verified, role, user_group, user_group_expire FROM users WHERE id = ? LIMIT 1", id)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var groupExpire int

	err = rows.Scan(&u.Id, &u.Username, &u.Email, &u.Password, &u.EmailVerified, &u.Role, &u.GroupId, &groupExpire)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	if groupExpire > 0 {
		u.GroupExpireTime.Valid = true
		u.GroupExpireTime.Time = time.Unix(int64(groupExpire), 0)
	}

	return &u, nil
}

func UserChangePassword(id int, password string) error {
	_, err := DB.Exec("UPDATE users SET password = ? WHERE id = ?", password, id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

// create a new user and return user id
func UserCreate(username, email, password string, email_verified bool, role int, group int) (int, error) {
	r, err := DB.Exec("INSERT INTO users(username, email, password, email_verified, role, user_group) VALUES (?, ?, ?, ?, ?, ?)", username, email, password, email_verified, role, group)
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}
	id, err := r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}
	return int(id), nil
}

func UserChangeUsername(id int, username string) error {
	_, err := DB.Exec("UPDATE users SET username = ? WHERE id = ?", username, id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

// update email and set email_verified to true
func UserChangeEmail(id int, email string) error {
	_, err := DB.Exec("UPDATE users SET email = ?, email_verified = ? WHERE id = ?", email, true, id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

// mark an email address as verified
func UserVerifyEmail(email string) error {
	_, err := DB.Exec("UPDATE users SET email_verified = ? WHERE email = ?", true, email)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

func UserFindAll(skip int, limit int) ([]User, error) {
	rows, err := DB.Query("SELECT id, username, email, password, email_verified, role, user_group, user_group_expire FROM users LIMIT ? OFFSET ?", limit, skip)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)

	for rows.Next() {
		var u User
		var groupExpire int

		err = rows.Scan(&u.Id, &u.Username, &u.Email, &u.Password, &u.EmailVerified, &u.Role, &u.GroupId, &groupExpire)
		if err != nil {
			return nil, fmt.Errorf("db: %w", err)
		}

		if groupExpire > 0 {
			u.GroupExpireTime.Valid = true
			u.GroupExpireTime.Time = time.Unix(int64(groupExpire), 0)
		}

		users = append(users, u)
	}

	return users, nil
}

func UserCount() (int, error) {
	row := DB.QueryRow("SELECT COUNT(*) FROM users")
	var cnt int
	err := row.Scan(&cnt)
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}
	return cnt, nil
}

func UserChangeRole(id int, role int) error {
	_, err := DB.Exec("UPDATE users SET role = ? WHERE id = ?", role, id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

func UserChangeGroup(id int, g int) error {
	_, err := DB.Exec("UPDATE users SET user_group = ? WHERE id = ?", g, id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

func UserChangeGroupExpire(id int, t int) error {
	_, err := DB.Exec("UPDATE users SET user_group_expire = ? WHERE id = ?", t, id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

// reset to default group for users with expired group membership
func UserResetExpiredGroup(id int) error {
	_, err := DB.Exec("UPDATE users SET user_group = ?, user_group_expire = 0 WHERE user_group_expire > 0 AND user_group_expire < unixepoch()", id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}
