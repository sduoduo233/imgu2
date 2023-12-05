package services

import (
	"bytes"
	"database/sql"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"img2/db"
	"img2/utils"
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
func (*upload) UploadImage(userId sql.NullInt32, file []byte, expire sql.NullTime, ipAddr string) (string, error) {
	contentType := http.DetectContentType(file)

	var encodedImage []byte // re-encoded image
	var fileExtension string

	if contentType == "image/gif" {

		// gif

		img, err := gif.DecodeAll(bytes.NewReader(file))
		if err != nil {
			return "", fmt.Errorf("decode gif: %w", err)
		}

		var buf bytes.Buffer

		err = gif.EncodeAll(&buf, img)
		if err != nil {
			return "", fmt.Errorf("encode gif: %w", err)
		}

		encodedImage = buf.Bytes()
		fileExtension = ".gif"

	} else {

		// png & jpeg

		img, format, err := image.Decode(bytes.NewReader(file))
		if err != nil {
			return "", fmt.Errorf("decode: %w", err)
		}

		var buf bytes.Buffer

		switch format {
		case "jpeg":
			fileExtension = ".jpg"
			err = jpeg.Encode(&buf, img, &jpeg.Options{
				Quality: 90,
			})
		case "png":
			fileExtension = ".png"
			err = png.Encode(&buf, img)
		default:
			return "", fmt.Errorf("unknown format: %s", format)
		}

		if err != nil {
			return "", fmt.Errorf("encode: %w", err)
		}

		encodedImage = buf.Bytes()

	}

	fileName := utils.RandomString(8) + fileExtension

	// upload file
	id, err := Storage.Put(fileName, encodedImage, expire)
	if err != nil {
		return "", err
	}

	// insert to database
	_, err = db.ImageCreate(id, userId, fileName, ipAddr, expire)
	if err != nil {
		return "", err
	}

	return fileName, nil

}
