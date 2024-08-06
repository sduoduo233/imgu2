package services

import (
	"fmt"
	"imgu2/db"
	"log/slog"
)

type image struct{}

var Image = image{}

// Get the content of the file using the public file name.
//
// return a byte array or a URL
//
// return nil if image not found
func (*image) Get(fileName string) (any, error) {
	img, err := db.ImageFindByFileName(fileName)
	if err != nil {
		return nil, err
	}

	if img == nil {
		return nil, nil
	}

	return Storage.GetFile(img.StorageId, img.InternalName)
}

// return nil if not found
func (*image) FindByFileName(fileName string) (*db.Image, error) {
	return db.ImageFindByFileName(fileName)
}

// find all images uploaded by a user
func (*image) FindByUser(userId int, page int) ([]db.Image, error) {
	const pageSize = 20
	i, err := db.ImageFindByUser(userId, page*pageSize, pageSize)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (*image) FindAll(page int) ([]db.Image, error) {
	const pageSize = 20
	i, err := db.ImageFindAll(page*pageSize, pageSize)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (*image) CountAll() (int, error) {
	return db.ImageCountAll()
}

func (*image) CountByUser(userId int) (int, error) {
	return db.ImageCountByUser(userId)
}

// permanently delete an image (delete from database and storage driver)
//
// error occurred when delete the file from storage is ignored if force == true
func (*image) Delete(i *db.Image, force bool) error {
	err := Storage.DeleteFileFromDriver(i.StorageId, i.InternalName)
	if err != nil {
		if force {
			slog.Error("delete image from storage", "file name", i.FileName, "internal name", i.InternalName, "storage", i.StorageId, "err", err)
		} else {
			return fmt.Errorf("delete image: %w", err)
		}
	}

	err = db.ImageDelete(i.Id)
	if err != nil {
		return fmt.Errorf("delete image: %w", err)
	}

	return nil
}
