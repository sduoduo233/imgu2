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
	"strings"
	"time"
)

func upload(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	maxTime, err := services.Upload.MaxUploadTime(user != nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("upload", "err", err)
		return
	}

	guestUpload, err := services.Setting.GetGuestUpload()
	if err != nil {
		slog.Error("upload", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	avifEnabled, err := services.Setting.IsAVIFEncodingEnabled()
	if err != nil {
		slog.Error("upload", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render(w, "upload", H{
		"user":         user,
		"csrf_token":   csrfToken(w),
		"max_time":     maxTime,
		"guest_upload": guestUpload,
		"avif_enabled": avifEnabled,
	})
}

func doUpload(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	// guest upload
	guestUpload, err := services.Setting.GetGuestUpload()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("do upload", "err", err)
		return
	}

	if user == nil && !guestUpload {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, H{
			"error": "GUEST_UPLOAD_NOT_ALLOWED",
		})
		return
	}

	// user role
	if user != nil && user.Role != services.RoleAdmin && user.Role != services.RoleUser {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, H{
			"error": "USER_BANNED",
		})
		return
	}

	// email verified
	if user != nil && !user.EmailVerified {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, H{
			"error": "EMAIL_NOT_VERIFIED",
		})
		return
	}

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

	// validate expire
	maxTime, err := services.Upload.MaxUploadTime(user != nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("do upload", "err", err)
		return
	}

	// limit exists && ( duration exceeds limit || never expire )
	if maxTime != 0 && ((expire > int(maxTime)) || (expire == 0)) {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, H{
			"error": "EXPIRE_TOO_LARGE",
		})
		return
	}

	file, fileHeaders, err := r.FormFile("file")
	if err != nil {
		slog.Debug("upload: read file", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// image format
	targetFormat := ""
	switch r.FormValue("format") {
	case "webp":
		targetFormat = "image/webp"
	case "jpeg":
		targetFormat = "image/jpeg"
	case "gif":
		targetFormat = "image/gif"
	case "png":
		targetFormat = "image/png"
	case "avif":
		targetFormat = "image/avif"

		avifEnabled, err := services.Setting.IsAVIFEncodingEnabled()
		if err != nil {
			slog.Error("upload", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !avifEnabled {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, H{
				"error": "UNSUPPORTED_ENCODING",
			})
			return
		}
	default:
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

	// encoding parameters
	var (
		lossless bool
		Q        int
		effort   int
	)

	lossless, err = strconv.ParseBool(r.FormValue("lossless"))
	if err != nil {
		slog.Error("do upload: parse encoding param 'lossless'", "err", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	Q, err = strconv.Atoi(r.FormValue("Q"))
	if err != nil {
		slog.Error("do upload: parse encoding param 'Q'", "err", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	effort, err = strconv.Atoi(r.FormValue("effort"))
	if err != nil {
		slog.Error("do upload: parse encoding param 'effort'", "err", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	// upload
	var fileName string
	if expire == 0 {
		fileName, err = services.Upload.UploadImage(userId, fileContent, sql.NullTime{}, ipAddr, targetFormat, int(maxSize), lossless, Q, effort)
	} else {
		t := time.Now().Add(time.Second * time.Duration(expire))
		fileName, err = services.Upload.UploadImage(userId, fileContent, sql.NullTime{Valid: true, Time: t}, ipAddr, targetFormat, int(maxSize), lossless, Q, effort)
	}

	if err != nil {
		slog.Error("do upload: upload", "err", err)
		w.WriteHeader(http.StatusInternalServerError)

		if strings.HasPrefix(err.Error(), "upload: ") {
			// storage driver error
			writeJSON(w, H{
				"error": "INTERNAL_STORAGE_ERROR",
			})
		} else {
			// malformated image
			writeJSON(w, H{
				"error": "IMAGE_PROCESSING_ERROR",
			})
		}
		return
	}

	writeJSON(w, H{
		"file_name": fileName,
	})
}
