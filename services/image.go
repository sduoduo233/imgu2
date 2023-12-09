package services

import "imgu2/db"

type image struct{}

var Image = image{}

// get the image from storage drivers
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

	return Storage.GetFile(img.StorageId, fileName)
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

func (*image) CountByUser(userId int) (int, error) {
	return db.ImageCountByUser(userId)
}
