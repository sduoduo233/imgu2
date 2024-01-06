package controllers

import (
	"imgu2/controllers/middleware"
	"imgu2/services"
	"log/slog"
	"net/http"
)

func adminSettings(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	m, err := services.Setting.GetAll()
	if err != nil {
		slog.Error("admin settings", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render(w, "admin_settings", H{
		"user":       user,
		"setting":    m,
		"csrf_token": csrfToken(w),
	})
}

func doAdminSettings(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	form := r.Form
	for k, v := range form {
		err := services.Setting.Set(k, v[0])
		if err != nil {
			slog.Error("admin settings", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	renderDialog(w, tr("info"), "Settings updated", "/admin/settings", tr("continue"))
}
