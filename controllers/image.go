package controllers

import (
	"imgu2/controllers/middleware"
	"imgu2/services"
	"imgu2/services/placeholder"
	"io"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"
)

func downloadImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET")

	fileName := chi.URLParam(r, "fileName")

	c, err := services.Image.Get(fileName)
	if err != nil {
		slog.Error("download image", "err", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("Content-Type", "image/png")
		w.Write(placeholder.ERROR)
		return
	}

	switch v := c.(type) {
	case string:
		http.Redirect(w, r, v, http.StatusFound)

	case []byte:
		w.Header().Add("Content-Type", http.DetectContentType(v))
		w.Write(v)

	case nil:
		w.Header().Add("Content-Type", "image/png")
		w.WriteHeader(http.StatusNotFound)
		w.Write(placeholder.NOT_FOUND)

	default:
		slog.Error("download image: unexpected type", "type", reflect.TypeOf(v))

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("Content-Type", "image/png")
		w.Write(placeholder.ERROR)
	}
}

func previewImage(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	fileName := chi.URLParam(r, "fileName")

	siteUrl, err := services.Setting.GetSiteURL()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("preview image", "err", err)
		return
	}

	img, err := services.Image.FindByFileName(fileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("preview image", "err", err)
		return
	}

	if img == nil {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "404 not found")
		return
	}

	// whether the current user is the owner of the image
	own := user != nil && img.Uploader.Valid && img.Uploader.Int32 == int32(user.Id)

	expire := int64(0)
	if img.ExpireTime.Valid {
		expire = img.ExpireTime.Time.Unix()
	}

	render(w, "preview", H{
		"user":        user,
		"file_name":   fileName,
		"site_url":    siteUrl,
		"uploaded_at": img.Time.Unix(),
		"expire":      expire,
		"own":         own,
		"csrf_token":  csrfToken(w),
	})
}
