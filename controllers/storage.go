package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"img2/controllers/middleware"
	"img2/services"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mattn/go-sqlite3"
)

func adminStorages(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	storages, err := services.Storage.FindAll()
	if err != nil {
		slog.Error("admin list storages", "err", err)
		renderDialog(w, "Error", "Unknown error", "/admin", "Go back")
		return
	}

	render(w, "storages", H{
		"user":     user,
		"storages": storages,
	})
}

func adminEditStorage(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s, err := services.Storage.FindById(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("edit storage", "err", err)
		return
	}

	config := make(map[string]string)

	err = json.Unmarshal([]byte(s.Config), &config)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("edit storage", "err", fmt.Errorf("unmarshal config: %w", err))
		return
	}

	render(w, "storage_config", H{
		"user":    user,
		"storage": s,
		"action":  "edit",
		"config":  config,
	})
}

func adminDoEditStorage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	form := r.Form
	enabled := form.Get("enabled") != ""
	allowUpload := form.Get("allow_upload") != ""

	config := make(map[string]string)

	for k, v := range form {
		key, ok := strings.CutPrefix(k, "config_")
		if !ok {
			continue
		}

		config[key] = v[0]
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		slog.Error("edit storage encode marshal json", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = services.Storage.Update(id, enabled, allowUpload, string(configJSON))
	if err != nil {
		slog.Error("edit storage update", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/storages", http.StatusFound)

}

func adminAddStorage(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	storageType := r.FormValue("type")

	if name == "" || storageType == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := services.Storage.Create(name, storageType)
	if err != nil {
		slog.Error("create storage", "err", err)

		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				// duplicated name
				renderDialog(w, "Error", "Duplicated storage name", "/admin/storages", "Go back")
				return
			}
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/storages/"+strconv.Itoa(id), http.StatusFound)
}

func adminStorageDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = services.Storage.Delete(id)
	if err != nil {
		slog.Error("delete storage", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/storages", http.StatusFound)
}
