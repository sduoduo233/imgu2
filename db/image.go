package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Image struct {
	Id         int
	StorageId  int
	Uploader   int
	FileName   string
	UploaderIP string
	Time       int64
	ExpireTime int64
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
