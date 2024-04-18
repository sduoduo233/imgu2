package services

import (
	"database/sql"
	"fmt"
	"imgu2/db"
	"imgu2/libvips"
)

type upload struct{}

var Upload = upload{}

// UploadImage re-encodes the image and save it to a random choosen storage driver
//
// userId may be set to nil to represent a guest user
//
// expire may be nil
//
// fileSizeLimit is the maximium file size in bytes after encoding
//
// return a random generated file name
func (*upload) UploadImage(userId sql.NullInt32, file []byte, expire sql.NullTime, ipAddr string, targetFormat string, fileSizeLimit int, lossless bool, Q int, effort int, contentType string) (string, error) {
	// re-encode image
	var fileExtension string

	// detect whether the source image is animated
	animated := false
	switch contentType {
	case "image/gif":
		animated = true
	case "image/webp":
		animated = true
	case "application/pdf":
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
	case "image/avif":
		fileExtension = ".avif"
		vipsForamt = libvips.FORMAT_AVIF
	default:
		return "", fmt.Errorf("upload: unknown format: %s", targetFormat)
	}

	encodedImage := libvips.LibvipsEncode(file, vipsForamt, animated, lossless, Q, effort)
	if encodedImage == nil {
		return "", fmt.Errorf("upload: malformatted image")
	}

	if len(encodedImage) > fileSizeLimit {
		return "", fmt.Errorf("upload: image too large")
	}

	fileName := RandomString(8) + fileExtension

	// upload file
	var err error
	var id int

	// internal name is the file name used in storage drivers
	internalName, id, err := Storage.Put(fileName, encodedImage, expire)
	if err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}

	// insert to database
	_, err = db.ImageCreate(id, userId, fileName, internalName, ipAddr, expire)
	if err != nil {
		return "", err
	}

	return fileName, nil

}
