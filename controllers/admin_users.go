package controllers

import (
	"fmt"
	"imgu2/controllers/middleware"
	"imgu2/services"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"
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

	groups, err := services.Group.FindAll()
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
		"groups":     groups,
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
		renderDialog(w, tr("error"), "An unknown error happened", "/admin/users", tr("go_back"))
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
}

func adminChangeUserGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	group, err := strconv.Atoi(r.FormValue("group"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = services.User.ChangeGroup(id, group)
	if err != nil {
		slog.Error("admin change group", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderDialog(w, tr("info"), tr("user_group_changed"), "/admin/users", tr("continue"))
}

func adminChangeUserGroupExpire(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	neverExpire := r.FormValue("never_expire") != ""
	if neverExpire {
		err := services.User.ChangeGroupExpire(id, 0)
		if err != nil {
			slog.Error("admin group exipre", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		t, err := time.Parse("2006-01-02 15:04 MST", fmt.Sprintf("%s %s UTC", r.FormValue("date"), r.FormValue("time")))
		if err != nil {
			slog.Error("admin group exipre", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = services.User.ChangeGroupExpire(id, int(t.Unix()))
		if err != nil {
			slog.Error("admin group exipre", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	renderDialog(w, tr("info"), tr("user_group_changed"), "/admin/users", tr("continue"))
}
