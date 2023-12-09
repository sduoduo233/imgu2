package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Image struct {
	Id         int
	StorageId  int
	Uploader   sql.NullInt32
	FileName   string
	UploaderIP string
	Time       time.Time
	ExpireTime sql.NullTime
}

// expire may be nil
//
// uploader may be nil to represent guest user
func ImageCreate(storage int, uploader sql.NullInt32, fileName string, uploaderIP string, expire sql.NullTime) (int, error) {

	// convert expire to unix time stamp
	expireUnix := sql.NullInt64{}
	if expire.Valid {
		// expireUnix is nil if expire is nil
		expireUnix.Valid = true
		expireUnix.Int64 = expire.Time.Unix()
	}

	r, err := DB.Exec("INSERT INTO images(storage, uploader, file_name, uploader_ip, time, expire_time) VALUES (?, ?, ?, ?, ?, ?)", storage, uploader, fileName, uploaderIP, time.Now().Unix(), expireUnix)
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	id, err := r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return int(id), nil
}

// find an image by file name
//
// return (nil, nil) if not found
func ImageFindByFileName(fileName string) (*Image, error) {
	var i Image
	var timeUnix int64
	var timeExpireUnix sql.NullInt64

	row := DB.QueryRow("SELECT id, storage, uploader, file_name, uploader_ip, time, expire_time FROM images WHERE file_name = ? AND (expire_time IS NULL OR expire_time > unixepoch()) LIMIT 1", fileName)
	err := row.Scan(&i.Id, &i.StorageId, &i.Uploader, &i.FileName, &i.UploaderIP, &timeUnix, &timeExpireUnix)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("db: %w", err)
	}

	i.Time = time.Unix(timeUnix, 0)

	if timeExpireUnix.Valid {
		i.ExpireTime.Valid = true
		i.ExpireTime.Time = time.Unix(timeExpireUnix.Int64, 0)
	}

	return &i, nil
}

func ImageFindExpired() ([]Image, error) {
	images := make([]Image, 0)

	rows, err := DB.Query("SELECT id, storage, uploader, file_name, uploader_ip, time, expire_time FROM images WHERE expire_time IS NOT NULL AND expire_time < unixepoch()")
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var i Image
		var timeUnix int64
		var timeExpireUnix sql.NullInt64
		err := rows.Scan(&i.Id, &i.StorageId, &i.Uploader, &i.FileName, &i.UploaderIP, &timeUnix, &timeExpireUnix)
		if err != nil {
			return nil, fmt.Errorf("db: %w", err)
		}

		i.Time = time.Unix(timeUnix, 0)

		if timeExpireUnix.Valid {
			i.ExpireTime.Valid = true
			i.ExpireTime.Time = time.Unix(timeExpireUnix.Int64, 0)
		}

		images = append(images, i)
	}

	return images, nil
}

func ImageDelete(id int) error {
	_, err := DB.Exec("DELETE FROM images WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

func ImageFindByUser(userId int, skip int, limit int) ([]Image, error) {
	images := make([]Image, 0)

	rows, err := DB.Query("SELECT id, storage, uploader, file_name, uploader_ip, time, expire_time FROM images WHERE uploader = ? AND (expire_time IS NULL OR expire_time > unixepoch()) LIMIT ? OFFSET ?", userId, limit, skip)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var i Image
		var timeUnix int64
		var timeExpireUnix sql.NullInt64
		err := rows.Scan(&i.Id, &i.StorageId, &i.Uploader, &i.FileName, &i.UploaderIP, &timeUnix, &timeExpireUnix)
		if err != nil {
			return nil, fmt.Errorf("db: %w", err)
		}

		i.Time = time.Unix(timeUnix, 0)

		if timeExpireUnix.Valid {
			i.ExpireTime.Valid = true
			i.ExpireTime.Time = time.Unix(timeExpireUnix.Int64, 0)
		}

		images = append(images, i)
	}

	return images, nil
}

// count uploaded images by a user
func ImageCountByUser(userId int) (int, error) {
	r := DB.QueryRow("SELECT COUNT(*) FROM images WHERE uploader = ? AND (expire_time IS NULL OR expire_time > unixepoch())", userId)

	var cnt int
	err := r.Scan(&cnt)
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}

	return cnt, nil
}
