package services

import (
	"database/sql"
	"fmt"
	"imgu2/db"
	"imgu2/libvips"
	"imgu2/utils"
	"net/http"
)

type upload struct{}

var Upload = upload{}

// UploadImage re-encodes the image and save it to a random choosen storage driver
//
// userId may be nil to represent a guest user
//
// expire may be nil
//
// return a random generated file name
func (*upload) UploadImage(userId sql.NullInt32, file []byte, expire sql.NullTime, ipAddr string, targetFormat string) (string, error) {
	// re-encode image
	var fileExtension string

	// detect whether the source image is animated
	animated := false
	contentType := http.DetectContentType(file)
	switch contentType {
	case "image/gif":
		animated = true
	case "image/webp":
		animated = true
	}

	var vipsForamt libvips.Format

	switch targetFormat {
	case "image/png":
		fileExtension = ".png"
		vipsForamt = libvips.FORMAT_PNG
	case "image/jpeg":
		fileExtension = ".jpg"
		vipsForamt = libvips.FORMAT_JEPG
	case "image/gif":
		fileExtension = ".gif"
		vipsForamt = libvips.FORMAT_GIF
	case "image/webp":
		fileExtension = ".webp"
		vipsForamt = libvips.FORMAT_WEBP
	default:
		return "", fmt.Errorf("upload: unknown format: %s", targetFormat)
	}

	encodedImage := libvips.LibvipsEncode(file, animated, vipsForamt)
	if encodedImage == nil {
		return "", fmt.Errorf("upload: malformatted image")
	}

	fileName := utils.RandomString(8) + fileExtension

	// upload file
	id, err := Storage.Put(fileName, encodedImage, expire)
	if err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}

	// insert to database
	_, err = db.ImageCreate(id, userId, fileName, ipAddr, expire)
	if err != nil {
		return "", err
	}

	return fileName, nil

}

// return maximum time in seconds an image is kept for
//
// return 0 if no duration limit is set
func (*upload) MaxUploadTime(login bool) (uint, error) {
	var maxTime uint = 0

	if !login {
		t, err := Setting.GetGuestUploadTime()
		if err != nil {
			return 0, err
		}
		if t > 0 {
			maxTime = t
		}
	} else {
		t, err := Setting.GetUserUploadTime()
		if err != nil {
			return 0, err
		}
		if t > 0 {
			maxTime = t
		}
	}

	return maxTime, nil
}
