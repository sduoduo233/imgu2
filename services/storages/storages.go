package storages

import (
	"database/sql"
)

type StorageDriver interface {
	// upload to a storage driver
	//
	// expire may be nil
	Put(key string, content []byte, expire sql.NullTime) error

	// delete a file from a storage driver
	Delete(key string) error

	// get a file from a storage driver
	//
	// return []byte which contains the content of the file
	// or a string which is a URL to the file
	Get(key string) (any, error)

	// ID returns the id of this dirver in database
	ID() int
}
