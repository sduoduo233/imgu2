package controllers

import (
	"errors"
	"fmt"
	"img2/controllers/middleware"
	"img2/services"
	"io"
	"log/slog"
	"net/http"

	"github.com/mattn/go-sqlite3"
)

func login(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUser(r.Context()) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	googleLogin, err := services.Setting.GetGoogleLogin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("login", "err", err)
		return
	}

	render(w, "login", H{
		"google_login": googleLogin,
	})
}

func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "TOKEN",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
	})
	renderDialog(w, "Info", "Logged out", "/login", "Login")
}

func doLogin(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUser(r.Context()) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	token, err := services.User.Login(email, password)
	if err != nil {
		renderDialog(w, "Error", "Incorrect email or password", "/login", "Go back")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "TOKEN",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func googleLogin(w http.ResponseWriter, r *http.Request) {
	googleLogin, err := services.Setting.GetGoogleLogin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("google login", "err", err)
		return
	}

	if !googleLogin {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Google login is disabled")
		return
	}

	u, err := services.User.GoogleSignin()
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		slog.Error("google login", "err", err)
		return
	}

	http.Redirect(w, r, u, http.StatusFound)
}

func googleLoginCallback(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	googleLogin, err := services.Setting.GetGoogleLogin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("google login", "err", err)
		return
	}

	if !googleLogin {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Google login is disabled")
		return
	}

	code := r.URL.Query().Get("code")

	if code == "" {
		renderDialog(w, "Error", "oauth error: "+r.URL.Query().Get("error"), "/login", "Go back")
		return
	}

	profile, err := services.User.GoogleCallback(code)

	if err != nil {
		slog.Error("google callback", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		renderDialog(w, "Error", "oauth error", "/login", "Go back")
		return
	}

	if user != nil {

		// already logged in
		err = services.User.LinkSocialAccount(services.SocialLoginGoogle, user.Id, profile)
		if err != nil {
			slog.Error("link google", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			renderDialog(w, "Error", "unknown error", "/dashboard/account", "Go back")
			return
		}

		http.Redirect(w, r, "/dashboard/account", http.StatusFound)
		return

	} else {

		// sign in or sign up with google

		token, err := services.User.SigninOrRegisterWithSocial(services.SocialLoginGoogle, profile)
		if err != nil {
			slog.Error("signin google", "err", err)

			var sqliteErr sqlite3.Error
			if errors.As(err, &sqliteErr) {
				if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
					// duplicated email
					renderDialog(w, "Error", fmt.Sprintf("An account with this email (%s) is already created. Please sign in to your original account.", profile.Email), "/login", "Go back")
					return
				}
			}

			w.WriteHeader(http.StatusInternalServerError)
			renderDialog(w, "Error", "unknown error", "/login", "Go back")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "TOKEN",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
		})

		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return

	}
}

func socialLoginUnlink(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	loginType := r.FormValue("type")
	if loginType != services.SocialLoginGoogle && loginType != services.SocialLoginGithub {
		w.WriteHeader(http.StatusBadRequest)
		renderDialog(w, "Error", "Bad request: invalid social login type", "", "")
		return
	}

	err := services.User.UnlinkSocialLogin(loginType, user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("unlink social account", "err", err)
		return
	}

	renderDialog(w, "Info", loginType+" account unlinked", "/dashboard/account", "Continue")
}
