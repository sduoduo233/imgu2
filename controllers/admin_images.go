package controllers

import (
	"imgu2/controllers/middleware"
	"imgu2/db"
	"imgu2/services"
	"log/slog"
	"math"
	"net/http"
	"strconv"
)

func adminImages(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		page = 0
	}

	// filters
	var images []db.Image
	var imageCount int

	uploader, err := strconv.Atoi(r.URL.Query().Get("uploader"))
	if err != nil || uploader < 0 {
		uploader = -1
	}

	if uploader < 0 { // filter is not set
		images, err = services.Image.FindAll(page)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("admin images", "err", err)
			return
		}

		imageCount, err = services.Image.CountAll()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("admin images", "err", err)
			return
		}
	} else {
		images, err = services.Image.FindByUser(uploader, page)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("admin images", "err", err)
			return
		}

		imageCount, err = services.Image.CountByUser(uploader)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("admin images", "err", err)
			return
		}
	}

	render(w, "admin_images", H{
		"csrf_token":      csrfToken(w),
		"user":            user,
		"images":          images,
		"page":            page,
		"total_page":      int(math.Ceil(float64(imageCount) / 20)),
		"filter_uploader": uploader,
	})
}

func adminImageDelete(w http.ResponseWriter, r *http.Request) {
	fileName := r.FormValue("file_name")
	if fileName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	img, err := services.Image.FindByFileName(fileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("admin image delete", "err", err)
		return
	}

	if img == nil {
		w.WriteHeader(http.StatusNotFound)
		renderDialog(w, "Error", "Image not found", "", "")
		return
	}

	err = services.Image.Delete(img)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("admin image delete", "err", err)
		return
	}

	renderDialog(w, "Info", "Image deleted", "/admin/images", "Go back")
}
