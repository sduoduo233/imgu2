package db

import "fmt"

type User struct {
	Id            int
	Username      string
	Email         string
	Password      string
	EmailVerified bool
	Role          int
	Space         int
}

func UserFindByEmail(email string) (*User, error) {
	var u User
	rows, err := DB.Query("SELECT id, username, email, password, email_verified, role, space FROM users WHERE email = ? LIMIT 1", email)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	err = rows.Scan(&u.Id, &u.Username, &u.Email, &u.Password, &u.EmailVerified, &u.Role, &u.Space)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return &u, nil
}

func UserFindById(id int) (*User, error) {
	var u User
	rows, err := DB.Query("SELECT id, username, email, password, email_verified, role, space FROM users WHERE id = ? LIMIT 1", id)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	err = rows.Scan(&u.Id, &u.Username, &u.Email, &u.Password, &u.EmailVerified, &u.Role, &u.Space)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
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
func UserCreate(username, email, password string, email_verified bool, role int) (int, error) {
	r, err := DB.Exec("INSERT INTO users(username, email, password, email_verified, role) VALUES (?, ?, ?, ?, ?)", username, email, password, email_verified, role)
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
