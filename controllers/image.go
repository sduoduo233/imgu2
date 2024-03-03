package controllers

import (
	"imgu2/controllers/middleware"
	"imgu2/services"
	"imgu2/services/placeholder"
	"io"
	"log/slog"
	"math"
	"net/http"
	"reflect"
	"strconv"

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
		w.Header().Add("Cache-Control", "no-cache")
		w.Header().Add("Content-Type", "image/png")
		w.Write(placeholder.ERROR)
		return
	}

	switch v := c.(type) {
	case string:
		http.Redirect(w, r, v, http.StatusMovedPermanently)

	case []byte:
		w.Header().Add("Content-Type", http.DetectContentType(v))
		w.Header().Add("Cache-Control", "max-age=31536000")
		w.Write(v)

	case nil: // not found
		w.Header().Add("Content-Type", "image/png")
		w.Header().Add("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusNotFound)
		w.Write(placeholder.NOT_FOUND)

	default:
		slog.Error("download image: unexpected type", "type", reflect.TypeOf(v))

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("Content-Type", "image/png")
		w.Header().Add("Cache-Control", "no-cache")
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

func myImages(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		page = 0
	}

	imageCount, err := services.Image.CountByUser(user.Id)
	if err != nil {
		slog.Error("my images", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	images, err := services.Image.FindByUser(user.Id, page)
	if err != nil {
		slog.Error("my images", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render(w, "images", H{
		"user":       user,
		"images":     images,
		"page":       page,
		"total_page": int(math.Ceil(float64(imageCount) / 20)), // page size = 20
	})
}

func deleteImage(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	fileName := r.FormValue("file_name")

	if fileName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	img, err := services.Image.FindByFileName(fileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("delete image", "err", err)
		return
	}

	if img == nil {
		w.WriteHeader(http.StatusNotFound)
		renderDialog(w, tr("error"), tr("image_not_found"), "/dashboard/images", tr("go_back"))
		return
	}

	// image uploaded by guest || the user is not the uploader
	if !img.Uploader.Valid || img.Uploader.Int32 != int32(user.Id) {
		w.WriteHeader(http.StatusForbidden)
		renderDialog(w, tr("error"), tr("no_permission_to_delete"), "/dashboard/images", tr("go_back"))
		return
	}

	err = services.Image.Delete(img)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		renderDialog(w, tr("error"), tr("unknown_error"), "/dashboard/images", tr("go_back"))
		slog.Error("delete image", "err", err)
		return
	}

	renderDialog(w, tr("info"), tr("image_deleted"), "/dashboard/images", tr("continue"))

}
