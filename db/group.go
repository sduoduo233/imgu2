package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
)

type Group struct {
	Id              int
	Name            string
	AllowUpload     bool
	MaxFileSize     int
	UploadPerMinute int
	UploadPerHour   int
	UploadPerDay    int
	UploadPerMonth  int
	TotalUpload     int

	// The number of seconds an uploaded image is kept for before it is deleted.
	// Zero means uploaded images are stored without a time limit.
	MaxRetentionSeconds int
}

// returns (nil, nil) if the group id does not exist
func GroupFindById(id int) (*Group, error) {
	var g Group

	row := DB.QueryRow("SELECT id, name, allow_upload, max_file_size, upload_per_minute, upload_per_hour, upload_per_day, upload_per_month, total_uploads, max_retention_seconds FROM groups WHERE id = ?", id)

	err := row.Scan(
		&g.Id,
		&g.Name,
		&g.AllowUpload,
		&g.MaxFileSize,
		&g.UploadPerMinute,
		&g.UploadPerHour,
		&g.UploadPerDay,
		&g.UploadPerMonth,
		&g.TotalUpload,
		&g.MaxRetentionSeconds,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("db: %w", err)
	}

	return &g, nil
}

func GroupFindAll() ([]Group, error) {

	groups := make([]Group, 0)

	rows, err := DB.Query("SELECT id, name, allow_upload, max_file_size, upload_per_minute, upload_per_hour, upload_per_day, upload_per_month, total_uploads, max_retention_seconds FROM groups")
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var g Group
		err := rows.Scan(
			&g.Id,
			&g.Name,
			&g.AllowUpload,
			&g.MaxFileSize,
			&g.UploadPerMinute,
			&g.UploadPerHour,
			&g.UploadPerDay,
			&g.UploadPerMonth,
			&g.TotalUpload,
			&g.MaxRetentionSeconds,
		)

		if err != nil {
			return nil, fmt.Errorf("db: %w", err)
		}

		groups = append(groups, g)
	}

	return groups, nil
}

// create a new user with random generated name
func GroupCreate() (int, error) {
	name := "New user group #" + strconv.Itoa(rand.Intn(10000))
	r, err := DB.Exec("INSERT INTO groups(name, max_file_size, upload_per_minute, upload_per_hour, upload_per_day, upload_per_month, total_uploads, allow_upload, max_retention_seconds) VALUES(?, 16000000, 30, 100, 1000, 1000, 10000, TRUE, 0)", name)

	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	id, err := r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return int(id), nil
}

// count the number of users in a user group
func GroupCountUsers(id int) (int, error) {
	row := DB.QueryRow("SELECT count(*) FROM users WHERE user_group = ?", id)

	var n int
	err := row.Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return n, nil
}

func GroupDelete(id int) error {
	_, err := DB.Exec("DELETE FROM groups WHERE id = ?", id)

	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func GroupEdit(id int, name string, allow_upload bool, max_file_size, upload_per_minute, upload_per_hour, upload_per_day, upload_per_month, total_uploads, max_retention_seconds int) error {
	_, err := DB.Exec("UPDATE groups SET name = ?, allow_upload = ?, max_file_size = ?, upload_per_minute = ?, upload_per_hour = ?, upload_per_day = ?, upload_per_month = ?, total_uploads = ?, max_retention_seconds = ? WHERE id = ?", name, allow_upload, max_file_size, upload_per_minute, upload_per_hour, upload_per_day, upload_per_month, total_uploads, max_retention_seconds, id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}
