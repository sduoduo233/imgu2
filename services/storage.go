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

// save the file in a randomly choosen storage driver
//
// return the file name and id of the storage driver
func (s *storage) Put(fileName string, content []byte, expire sql.NullTime) (string, int, error) {
	if len(s.uploadDrivers) == 0 {
		return "", 0, fmt.Errorf("no storage driver available")
	}

	n := RandomNumber(0, len(s.uploadDrivers))
	d := s.uploadDrivers[n]

	newFileName, err := d.Put(fileName, content, expire)
	if err != nil {
		return "", 0, fmt.Errorf("storage put: %w", err)
	}

	if newFileName != "" {
		fileName = newFileName
	}

	slog.Debug("put file", "file name", fileName, "size", len(content), "expire", expire)

	return fileName, d.ID(), nil
}

func (s *storage) DeleteFileFromDriver(id int, fileName string) error {
	slog.Debug("delete from driver", "id", id, "file name", fileName)

	for _, v := range s.dirvers {
		if v.ID() == id {
			return v.Delete(fileName)
		}
	}

	return fmt.Errorf("storage driver %d does not exist", id)
}

func (s *storage) GetFile(id int, fileName string) (any, error) {
	for _, v := range s.dirvers {
		if v.ID() == id {
			return v.Get(fileName)
		}
	}

	return nil, fmt.Errorf("storage driver %d does not exist", id)
}
