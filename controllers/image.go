package controllers

import (
	"imgu2/controllers/middleware"
	"imgu2/services"
	"io"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"
)

func downloadImage(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "fileName")

	c, err := services.Image.Get(fileName)
	if err != nil {
		slog.Error("download image", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "failed to load image")
		return
	}

	switch v := c.(type) {
	case string:
		http.Redirect(w, r, v, http.StatusFound)

	case []byte:
		w.Header().Add("Content-Type", http.DetectContentType(v))
		w.Write(v)

	case nil:
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "image not found")

	default:
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("download image: unexpected type", "type", reflect.TypeOf(v))
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

	render(w, "preview", H{
		"user":      user,
		"file_name": fileName,
		"site_url":  siteUrl,
	})
}
