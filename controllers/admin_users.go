package controllers

import (
	"imgu2/controllers/middleware"
	"imgu2/services"
	"log/slog"
	"math"
	"net/http"
	"strconv"
)

func adminUsers(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		page = 0
	}

	userCount, err := services.User.CountAll()
	if err != nil {
		slog.Error("admin users", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	users, err := services.User.FindAll(page)
	if err != nil {
		slog.Error("admin users", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render(w, "admin_users", H{
		"user":       user,
		"users":      users,
		"page":       page,
		"total_page": int(math.Ceil(float64(userCount) / 100)),
		"csrf_token": csrfToken(w),
	})
}

func adminChangeUserRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	role, err := strconv.Atoi(r.FormValue("role"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = services.User.ChangeRole(id, role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("admin change role", "err", err)
		renderDialog(w, "Error", "An unknown error happened", "/admin/users", "Go back")
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
}
