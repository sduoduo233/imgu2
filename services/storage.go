package services

import (
	"database/sql"
	"fmt"
	"imgu2/db"
	"imgu2/services/storages"
	"log/slog"
)

type storage struct {
	// all initialized storage drivers
	dirvers []storages.StorageDriver

	// storage drivers that has upload enableds
	uploadDrivers []storages.StorageDriver
}

var Storage = storage{}

type StorageType string

const (
	StorageLocal     StorageType = "local"
	StorageS3        StorageType = "s3"
	StorageFTP       StorageType = "ftp"
	StorageWebDAV    StorageType = "webdav"
	StorageTelegraph StorageType = "telegraph"
)

// initialize storage drivers
func (s *storage) Init() error {
	all, err := db.StorageFindAll()
	if err != nil {
		return err
	}

	for _, v := range all {
		if !v.Enabled {
			continue
		}

		var driver storages.StorageDriver

		switch v.Type {
		case string(StorageLocal):
			driver, err = storages.NewLocalStorage(v.Name, v.Id, v.Config)

		case string(StorageS3):
			driver, err = storages.NewS3Storage(v.Name, v.Id, v.Config)

		case string(StorageFTP):
			driver, err = storages.NewFTPStorage(v.Name, v.Id, v.Config)

		case string(StorageWebDAV):
			driver, err = storages.NewWebDAVStorage(v.Name, v.Id, v.Config)

		case string(StorageTelegraph):
			driver, err = storages.NewTelegraphStorage(v.Name, v.Id, v.Config)

		default:
			slog.Error("unknown storage type", "type", v.Type)
		}

		if err != nil {
			slog.Error("storage driver disabled due to initialization failure", "err", err)
			err = Storage.SetEnabled(v.Id, false)
			if err != nil {
				slog.Error("failed to disable storage driver", "err", err)
			}
			continue
		}

		s.dirvers = append(s.dirvers, driver)
		if v.AllowUpload {
			s.uploadDrivers = append(s.uploadDrivers, driver)
		}
	}

	slog.Info("storage drivers initialized", "count", len(s.dirvers))
	return nil
}

func (*storage) SetEnabled(id int, enabled bool) error {
	return db.StorageSetEnabled(id, enabled)
}

func (*storage) FindAll() ([]db.Storage, error) {
	return db.StorageFindAll()
}

func (*storage) FindById(id int) (*db.Storage, error) {
	return db.StorageFindById(id)
}

func (*storage) Update(id int, enabled bool, allowUpload bool, config string) error {
	return db.StorageUpdate(id, enabled, allowUpload, config)
}

func (*storage) Create(name string, t string) (int, error) {
	if t != string(StorageS3) && t != string(StorageLocal) && t != string(StorageFTP) && t != string(StorageWebDAV) && t != string(StorageTelegraph) {
		return 0, fmt.Errorf("unknown storage type: %s", t)
	}
	return db.StorageCreate(name, t, "{}", false, false)
}

func (*storage) Delete(id int) error {
	cnt, err := db.ImageCountByStorage(id)
	if err != nil {
		return err
	}

	if cnt > 0 {
		return fmt.Errorf("non empty storage driver can not be deleted")
	}

	return db.StorageDelete(id)
}

// Put uploads the file to a random choosen storage driver.
// Put may use the internalName supplied if the storage driver allows custom names.
//
// return the file name and the storage driver id
func (s *storage) Put(internalName string, content []byte, expire sql.NullTime) (string, int, error) {
	if len(s.uploadDrivers) == 0 {
		return "", 0, fmt.Errorf("no storage driver available")
	}

	n := RandomNumber(0, len(s.uploadDrivers))
	d := s.uploadDrivers[n]

	newFileName, err := d.Put(internalName, content, expire)
	if err != nil {
		return "", 0, fmt.Errorf("storage put: %w", err)
	}

	// Some storage driver does not allow setting file name, so a
	// new file name may be returned.
	if newFileName != "" {
		internalName = newFileName
	}

	slog.Debug("put file", "internal name", internalName, "size", len(content), "expire", expire)

	return internalName, d.ID(), nil
}

func (s *storage) DeleteFileFromDriver(id int, internalName string) error {
	slog.Debug("delete from driver", "id", id, "file name", internalName)

	for _, v := range s.dirvers {
		if v.ID() == id {
			return v.Delete(internalName)
		}
	}

	return fmt.Errorf("storage driver %d does not exist", id)
}

func (s *storage) GetFile(id int, internalName string) (any, error) {
	for _, v := range s.dirvers {
		if v.ID() == id {
			return v.Get(internalName)
		}
	}

	return nil, fmt.Errorf("storage driver %d does not exist", id)
}
