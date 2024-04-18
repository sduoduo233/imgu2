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

	avifEnabled, err := services.Setting.IsAVIFEncodingEnabled()
	if err != nil {
		slog.Error("upload", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	webpEnabled, err := services.Setting.IsWEBPEncodingEnabled()
	if err != nil {
		slog.Error("upload", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	group, err := services.Group.GetUserGroup(user)
	if err != nil {
		slog.Error("upload", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render(w, "upload", H{
		"user":         user,
		"csrf_token":   csrfToken(w),
		"max_time":     group.MaxRetentionSeconds,
		"group":        group,
		"avif_enabled": avifEnabled,
		"webp_enabled": webpEnabled,
	})
}

func doUpload(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	// disallow banned users from uploading
	if user != nil && user.Role == services.RoleBanned {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, H{
			"error": "USER_BANNED",
		})
		return
	}

	// disallow unverified users from uploading
	if user != nil && !user.EmailVerified {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, H{
			"error": "EMAIL_NOT_VERIFIED",
		})
		return
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

	// check user group policies
	group, err := services.Group.GetUserGroup(user)
	if err != nil {
		slog.Error("upload", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !group.AllowUpload {
		// uploading is disabled for this user group
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, H{
			"error": "PERMISSION_DENIED",
		})
		return
	}

	// image retention seconds limit
	if group.MaxRetentionSeconds != 0 {

		// expire == 0:
		// the user requests to store the image indefinitely while
		// there is a limit

		// group.MaxRetentionSeconds < expire:
		// the requested time is larger than the allowed value

		if group.MaxRetentionSeconds < expire || expire == 0 {
			w.WriteHeader(http.StatusForbidden)
			writeJSON(w, H{
				"error": "EXPIRE_TOO_LARGE",
			})
			return
		}
	}

	file, fileHeaders, err := r.FormFile("file")
	if err != nil {
		slog.Debug("upload: read file", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// file size limit
	if fileHeaders.Size > int64(group.MaxFileSize) {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, H{
			"error": "FILE_TOO_LARGE",
		})
		return
	}

	// image format
	targetFormat := ""
	switch r.FormValue("format") {
	case "webp":
		targetFormat = "image/webp"

		webpEnabled, err := services.Setting.IsWEBPEncodingEnabled()
		if err != nil {
			slog.Error("upload", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !webpEnabled {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, H{
				"error": "UNSUPPORTED_ENCODING",
			})
			return
		}

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
		return
	}
	Q, err = strconv.Atoi(r.FormValue("Q"))
	if err != nil {
		slog.Error("do upload: parse encoding param 'Q'", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	effort, err = strconv.Atoi(r.FormValue("effort"))
	if err != nil {
		slog.Error("do upload: parse encoding param 'effort'", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId := sql.NullInt32{}
	if user != nil {
		userId.Valid = true
		userId.Int32 = int32(user.Id)
	}

	// upload
	var fileName string
	if expire == 0 {
		fileName, err = services.Upload.UploadImage(userId, fileContent, sql.NullTime{}, ipAddr, targetFormat, group.MaxFileSize, lossless, Q, effort, fileHeaders.Header.Get("Content-Type"))
	} else {
		t := time.Now().Add(time.Second * time.Duration(expire))
		fileName, err = services.Upload.UploadImage(userId, fileContent, sql.NullTime{Valid: true, Time: t}, ipAddr, targetFormat, group.MaxFileSize, lossless, Q, effort, fileHeaders.Header.Get("Content-Type"))
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
