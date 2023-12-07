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
