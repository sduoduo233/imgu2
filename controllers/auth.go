package controllers

import (
	"errors"
	"fmt"
	"imgu2/controllers/middleware"
	"imgu2/services"
	"io"
	"log/slog"
	"net/http"
	"strings"

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

	githubLogin, err := services.Setting.GetGithubLogin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("login", "err", err)
		return
	}

	render(w, "login", H{
		"google_login": googleLogin,
		"github_login": githubLogin,
		"csrf_token":   csrfToken(w),
	})
}

func logout(w http.ResponseWriter, r *http.Request) {
	setCookie(w, "TOKEN", "")
	renderDialog(w, tr("info"), tr("logged_out"), "/login", tr("login"))
}

func doLogin(w http.ResponseWriter, r *http.Request) {
	if middleware.GetUser(r.Context()) != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	email := strings.ToLower(r.FormValue("email"))
	password := r.FormValue("password")

	token, err := services.Auth.Login(email, password)
	if err != nil {
		renderDialog(w, tr("error"), tr("incorrect_email_or_password"), "/login", tr("go_back"))
		return
	}

	setCookie(w, "TOKEN", token)

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// github login
func githubLogin(w http.ResponseWriter, r *http.Request) {
	g := services.Auth.GithubOAuth()
	if g == nil {
		io.WriteString(w, "github login is disabled")
		return
	}

	u, err := g.RedirectLink()
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		slog.Error("github login", "err", err)
		return
	}

	http.Redirect(w, r, u, http.StatusFound)
}

// google login
func googleLogin(w http.ResponseWriter, r *http.Request) {
	g := services.Auth.GoogleOAuth()
	if g == nil {
		io.WriteString(w, "google login is disabled")
		return
	}

	u, err := g.RedirectLink()
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		slog.Error("google login", "err", err)
		return
	}

	http.Redirect(w, r, u, http.StatusFound)
}

// google login callback
func googleLoginCallback(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	g := services.Auth.GoogleOAuth()
	if g == nil {
		io.WriteString(w, "google login is disabled")
		return
	}

	code := r.URL.Query().Get("code")

	if code == "" {
		renderDialog(w, tr("error"), tr("oauth_error")+": "+r.URL.Query().Get("error"), "/login", tr("go_back"))
		return
	}

	profile, err := g.GetProfile(code)

	if err != nil {
		slog.Error("google callback", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		renderDialog(w, tr("error"), tr("oauth_error"), "/login", tr("go_back"))
		return
	}

	if user != nil {

		// already logged in
		// link google account to an existing account
		err = services.Auth.LinkSocialAccount(services.SocialLoginGoogle, user.Id, profile)
		if err != nil {
			slog.Error("link google", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			renderDialog(w, tr("error"), tr("unknown_error"), "/dashboard/account", tr("go_back"))
			return
		}

		http.Redirect(w, r, "/dashboard/account", http.StatusFound)
		return

	} else {

		// sign in or sign up with google

		token, err := services.Auth.SigninOrRegisterWithSocial(services.SocialLoginGoogle, profile)
		if err != nil {
			var sqliteErr sqlite3.Error
			if errors.As(err, &sqliteErr) {
				if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
					// duplicated email
					renderDialog(w, tr("error"), fmt.Sprintf(tr("oauth_dup_email"), profile.Email), "/login", tr("go_back"))
					return
				}
			}

			if err.Error() == "registration is disabled" {
				renderDialog(w, tr("error"), tr("registration_disabled"), "/login", tr("go_back"))
				return
			}

			slog.Error("signin google", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			renderDialog(w, tr("error"), tr("unknown_error"), "/login", tr("go_back"))
			return
		}

		setCookie(w, "TOKEN", token)

		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return

	}
}

// github login callback
func githubLoginCallback(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())

	g := services.Auth.GithubOAuth()
	if g == nil {
		io.WriteString(w, "github login is disabled")
		return
	}

	code := r.URL.Query().Get("code")

	if code == "" {
		renderDialog(w, tr("error"), tr("oauth_error")+": "+r.URL.Query().Get("error"), "/login", tr("go_back"))
		return
	}

	// get github user profile
	profile, err := g.GetProfile(code)
	if err != nil {
		slog.Error("github callback", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		renderDialog(w, tr("error"), tr("oauth_error"), "/login", tr("go_back"))
		return
	}

	if user != nil {

		// already logged in
		// link github account to an existing account
		err = services.Auth.LinkSocialAccount(services.SocialLoginGithub, user.Id, profile)
		if err != nil {
			slog.Error("link github", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			renderDialog(w, tr("error"), tr("unknown_error"), "/dashboard/account", tr("go_back"))
			return
		}

		http.Redirect(w, r, "/dashboard/account", http.StatusFound)
		return

	} else {

		// sign in or sign up with github

		token, err := services.Auth.SigninOrRegisterWithSocial(services.SocialLoginGithub, profile)
		if err != nil {

			var sqliteErr sqlite3.Error
			if errors.As(err, &sqliteErr) {
				if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
					// duplicated email
					renderDialog(w, tr("error"), fmt.Sprintf(tr("oauth_dup_email"), profile.Email), "/login", tr("go_back"))
					return
				}
			}

			if err.Error() == "registration is disabled" {
				renderDialog(w, tr("error"), tr("registration_disabled"), "/login", tr("go_back"))
				return
			}

			slog.Error("signin github", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			renderDialog(w, tr("error"), "unknown error", "/login", tr("go_back"))
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
		renderDialog(w, tr("error"), "Bad request: invalid social login type", "", "")
		return
	}

	err := services.Auth.UnlinkSocialLogin(loginType, user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("unlink social account", "err", err)
		return
	}

	renderDialog(w, tr("info"), loginType+tr("oauth_account_unlinked"), "/dashboard/account", tr("continue"))
}
