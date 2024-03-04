package controllers

import (
	"errors"
	"imgu2/controllers/middleware"
	"imgu2/services"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/mattn/go-sqlite3"
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

	googleLinked, err := services.Auth.SocialLoginLinked(services.SocialLoginGoogle, user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("account setting", "err", err)
		return
	}

	githubLogin, err := services.Setting.GetGithubLogin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("account setting", "err", err)
		return
	}

	githubLinked, err := services.Auth.SocialLoginLinked(services.SocialLoginGithub, user.Id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("account setting", "err", err)
		return
	}

	render(w, "account", H{
		"user":          user,
		"google_login":  googleLogin,
		"google_linked": googleLinked,
		"github_login":  githubLogin,
		"github_linked": githubLinked,
		"csrf_token":    csrfToken(w),
	})
}

func changePassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	current := r.FormValue("current")
	password1 := r.FormValue("password1")
	password2 := r.FormValue("password2")

	if password1 != password2 {
		w.WriteHeader(http.StatusBadRequest)
		renderDialog(w, tr("error"), tr("password_do_not_match"), "", "")
		return
	}

	if len(password1) > 30 || len(password1) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		renderDialog(w, tr("error"), tr("password_wrong_length"), "", "")
		return
	}

	err := services.User.ChangePassword(user.Id, current, password1)
	if err != nil {
		if err.Error() == "current password does not match" {
			renderDialog(w, tr("error"), tr("wrong_current_password"), "", "")
			return
		}
		slog.Error("change password", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderDialog(w, tr("info"), tr("password_updated"), "/dashboard", tr("continue"))
}

func changeUsername(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	username := r.FormValue("username")
	match, err := regexp.Match("^[\\w #]{3,30}$", []byte(username))
	if err != nil || !match {
		renderDialog(w, tr("error"), tr("invalid_username"), "/register", tr("go_back"))
		return
	}

	err = services.User.ChangeUsername(user.Id, username)
	if err != nil {
		slog.Error("change username", "err", err)

		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			// duplicated username
			renderDialog(w, tr("error"), tr("dup_username"), "/dashboard/account", tr("go_back"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderDialog(w, tr("info"), tr("username_changed"), "/dashboard/account", tr("continue"))
}

func changeEmail(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	email := strings.ToLower(r.FormValue("email"))

	match, err := regexp.Match("^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$", []byte(email))
	if err != nil || !match {
		renderDialog(w, tr("error"), tr("invalid_email"), "/register", tr("go_back"))
		return
	}

	err = services.User.ChangeEmail(user.Id, email)
	if err != nil {
		slog.Error("change email", "err", err)

		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			// duplicated email
			renderDialog(w, tr("error"), tr("dup_email"), "/dashboard/account", tr("go_back"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderDialog(w, tr("info"), tr("change_email_verification_email_sent"), "/dashboard/account", tr("continue"))
}

func verifyEmail(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	if user.EmailVerified {
		io.WriteString(w, "Your email is already verified")
		return
	}

	render(w, "verify_email", H{
		"user":       user,
		"csrf_token": csrfToken(w),
	})
}

func doVerifyEmail(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	if user.EmailVerified {
		io.WriteString(w, "Your email is already verified")
		return
	}

	err := services.User.SendVerificationEmail(user.Id)
	if err != nil {
		slog.Error("send verification email", "err", err)
		renderDialog(w, tr("error"), "Unknown error", "/dashboard", tr("go_back"))
		return
	}

	renderDialog(w, tr("info"), tr("verification_email_sent"), "/dashboard", tr("continue"))
}

func verifyEmailCallback(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		renderDialog(w, tr("error"), "Token is empty", "", "")
		return
	}

	err := services.User.VerifyEmail(token)
	if err != nil {
		slog.Error("verify email", "err", err)
		renderDialog(w, tr("error"), "Invalid token", "", "")
		return
	}

	renderDialog(w, tr("info"), tr("email_verified"), "/dashboard", tr("continue"))
}

func changeEmailCallback(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		renderDialog(w, tr("error"), "Token is empty", "", "")
		return
	}

	err := services.User.ChangeEmailCallback(token)
	if err != nil {
		slog.Error("change email callback", "err", err)
		renderDialog(w, tr("error"), "Invalid token", "", "")
		return
	}

	renderDialog(w, tr("info"), tr("email_changed"), "/dashboard", tr("continue"))
}

func register(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	render(w, "register", H{
		"csrf_token": csrfToken(w),
	})
}

func doRegister(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r.Context())
	if user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	allowRegister, err := services.Setting.GetAllowRegister()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("register", "err", err)
		return
	}

	if !allowRegister {
		renderDialog(w, tr("error"), tr("registration_disabled"), "/login", tr("go_back"))
		return
	}

	username := r.FormValue("username")
	email := strings.ToLower(r.FormValue("email"))
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	if username == "" || email == "" || password == "" || password2 == "" {
		w.WriteHeader(http.StatusBadRequest)
		renderDialog(w, tr("error"), tr("missing_required_fields"), "/register", tr("go_back"))
		return
	}

	if password != password2 {
		renderDialog(w, tr("error"), tr("password_do_not_match"), "/register", tr("go_back"))
		return
	}

	// email check
	match, err := regexp.Match("^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$", []byte(email))
	if err != nil || !match {
		renderDialog(w, tr("error"), tr("invalid_email"), "/register", tr("go_back"))
		return
	}

	// username check (3-30 characters, alphanumeric & underscore, #, space)
	match, err = regexp.Match("^[\\w #]{3,30}$", []byte(username))
	if err != nil || !match {
		renderDialog(w, tr("error"), tr("invalid_username"), "/register", tr("go_back"))
		return
	}

	// password check
	if len(password) > 30 || len(password) < 8 {
		renderDialog(w, tr("error"), tr("password_wrong_length"), "/register", tr("go_back"))
		return
	}

	token, err := services.User.Register(username, email, password)
	if err != nil {
		slog.Error("register", "err", err)

		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			// duplicated username / email
			renderDialog(w, tr("error"), tr("dup_username_or_email"), "/register", tr("go_back"))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		renderDialog(w, tr("error"), tr("unknown_error"), "/register", tr("go_back"))
		return
	}

	setCookie(w, "TOKEN", token)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func resetPasswordCallback(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	render(w, "reset_password_callback", H{
		"csrf_token": csrfToken(w),
		"token":      token,
	})
}

func doResetPasswordCallback(w http.ResponseWriter, r *http.Request) {
	password1 := r.FormValue("password")
	password2 := r.FormValue("password2")

	if password1 != password2 {
		renderDialog(w, tr("error"), tr("password_do_not_match"), "", "")
		return
	}

	if len(password1) > 30 || len(password1) < 8 {
		renderDialog(w, tr("error"), tr("password_wrong_length"), "", "")
		return
	}

	err := services.User.ResetPasswordCallback(r.FormValue("token"), password1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("reset password callback", "err", err)
		renderDialog(w, tr("error"), tr("unknown_error"), "", "")
		return
	}

	renderDialog(w, tr("info"), tr("password_updated"), "/login", "Login")
}

func resetPassword(w http.ResponseWriter, r *http.Request) {
	render(w, "reset_password", H{
		"csrf_token": csrfToken(w),
	})
}

func doResetPassword(w http.ResponseWriter, r *http.Request) {
	email := strings.ToLower(r.FormValue("email"))
	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := services.User.ResetPassword(email)
	if err != nil {
		if err.Error() == "email unverified" {
			renderDialog(w, tr("error"), tr("email_unverified"), "/login", tr("go_back"))
			return
		}

		slog.Error("reset password", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		renderDialog(w, tr("error"), tr("unknown_error"), "/login", tr("go_back"))
		return
	}

	renderDialog(w, tr("info"), tr("reset_password_email_sent"), "/login", "OK")
}
