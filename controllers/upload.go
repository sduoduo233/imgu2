package controllers

import (
	"database/sql"
	"imgu2/controllers/middleware"
	"imgu2/services"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

func upload(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	render(w, "upload", H{
		"user": user,
	})
}

func doUpload(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	userId := sql.NullInt32{}
	if user != nil {
		userId.Valid = true
		userId.Int32 = int32(user.Id)
	}

	// exipre in seconds
	expire, err := strconv.Atoi(r.FormValue("expire"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if expire < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, fileHeaders, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check file size
	maxSize, err := services.Setting.GetMaxImageSize()
	if err != nil {
		slog.Error("do upload", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if fileHeaders.Size > int64(maxSize) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		writeJSON(w, H{
			"error": "FILE_TOO_LARGE",
		})
		return
	}

	// read uploaded file
	fileContent, err := io.ReadAll(file)
	if err != nil {
		slog.Error("do upload: read file", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ip address
	ipAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("do upload: remote addr", "err", err)
		return
	}

	// upload
	var fileName string
	if expire == 0 {
		fileName, err = services.Upload.UploadImage(userId, fileContent, sql.NullTime{}, ipAddr)
	} else {
		t := time.Now().Add(time.Second * time.Duration(expire))
		fileName, err = services.Upload.UploadImage(userId, fileContent, sql.NullTime{Valid: true, Time: t}, ipAddr)
	}

	if err != nil {
		slog.Error("do upload: upload", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, H{
			"error": "IMAGE_PROCESSING_ERROR",
		})
		return
	}

	writeJSON(w, H{
		"file_name": fileName,
	})
}
