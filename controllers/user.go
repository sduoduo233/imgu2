package controllers

import (
	"img2/controllers/middleware"
	"img2/services"
	"log/slog"
	"net/http"
)

func dashboardIndex(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	render(w, "dashboard", H{
		"user": user,
	})
}

func accountSetting(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	googleLogin, err := services.Setting.GetGoogleLogin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("account setting", "err", err)
		return
	}

	googleLinked, err := services.User.SocialLoginLinked(services.SocialLoginGoogle, user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("account setting", "err", err)
		return
	}

	render(w, "account", H{
		"user":          user,
		"google_login":  googleLogin,
		"google_linked": googleLinked,
	})
}

func changePassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	current := r.FormValue("current")
	password1 := r.FormValue("password1")
	password2 := r.FormValue("password2")

	if password1 != password2 {
		w.WriteHeader(http.StatusBadRequest)
		renderDialog(w, "Error", "Password does not match", "", "")
		return
	}

	if len(password1) > 32 || len(password1) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		renderDialog(w, "Error", "Password must be between 8 to 32 characters", "", "")
		return
	}

	err := services.User.ChangePassword(user.Id, current, password1)
	if err != nil {
		if err.Error() == "current password does not match" {
			renderDialog(w, "Error", "Current password does not match", "", "")
			return
		}
		slog.Error("change password", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderDialog(w, "Info", "Password updated", "/dashboard", "Continue")
	return
}
