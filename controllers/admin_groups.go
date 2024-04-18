package controllers

import (
	"imgu2/controllers/middleware"
	"imgu2/services"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// list all user groups
func adminGroups(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	groups, err := services.Group.FindAll()
	if err != nil {
		slog.Error("admin groups", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render(w, "admin_groups", H{
		"user":       user,
		"groups":     groups,
		"csrf_token": csrfToken(w),
	})
}

// display the edit page
func adminGroupEdit(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	g, err := services.Group.FindById(id)
	if err != nil {
		slog.Error("admin group edit", "err", err, "id", id)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if g == nil {
		w.WriteHeader(http.StatusNotFound)
		renderDialog(w, tr("error"), tr("group_not_found"), "", "")
		return
	}

	render(w, "admin_groups_edit", H{
		"csrf_token": csrfToken(w),
		"group":      g,
		"user":       user,
	})

}

// edit user group policies
func adminGroupDoEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	atoi := func(s string) int {
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0 // ignore parsing error
		}
		return i
	}

	err = services.Group.Edit(
		id,
		r.FormValue("name"),
		r.FormValue("allow_upload") != "",
		atoi(r.FormValue("max_file_size")),
		atoi(r.FormValue("upload_per_minute")),
		atoi(r.FormValue("upload_per_hour")),
		atoi(r.FormValue("upload_per_day")),
		atoi(r.FormValue("upload_per_month")),
		atoi(r.FormValue("total_uploads")),
		atoi(r.FormValue("max_retention_seconds")),
	)

	if err != nil {
		slog.Error("admin group edit", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderDialog(w, tr("info"), tr("group_edited"), "/admin/groups", "Go back")
}

func adminGroupCreate(w http.ResponseWriter, r *http.Request) {
	id, err := services.Group.Create()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("admin gruop create", "err", err)
		renderDialog(w, tr("error"), tr("error_create_group"), "/admin/groups", "Go back")
		return
	}

	renderDialog(w, tr("info"), tr("group_created"), "/admin/groups/"+strconv.Itoa(id), "Continue")
}

func adminGroupDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cnt, err := services.Group.CountUsers(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("admin group delete", "err", err)
		return
	}

	if cnt > 0 {
		renderDialog(w, tr("error"), tr("group_not_empty"), "/admin/groups", "Go back")
		return
	}

	err = services.Group.Delete(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("admin group delete", "err", err)
		return
	}

	renderDialog(w, tr("info"), tr("group_deleted"), "/admin/groups", "Continue")
}
