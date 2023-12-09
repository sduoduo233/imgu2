package controllers

import (
	"errors"
	"imgu2/controllers/middleware"
	"imgu2/services"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
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

	googleLinked, err := services.User.SocialLoginLinked(services.SocialLoginGoogle, user.Id)
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

	githubLinked, err := services.User.SocialLoginLinked(services.SocialLoginGithub, user.Id)
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
}

func changeUsername(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	username := r.FormValue("username")
	if len(username) < 5 || len(username) > 30 {
		renderDialog(w, "Error", "Username must be between 5 to 30 characters", "/dashboard/account", "Go back")
		return
	}

	err := services.User.ChangeUsername(user.Id, username)
	if err != nil {
		slog.Error("change username", "err", err)

		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			// duplicated username
			renderDialog(w, "Error", "This username is already used by another user.", "/dashboard/account", "Go back")
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderDialog(w, "Error", "Username changed", "/dashboard/account", "Continue")
}

func changeEmail(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	email := r.FormValue("email")

	if len(email) == 0 || len(email) > 50 || strings.Count(email, "@") != 1 {
		renderDialog(w, "Error", "Invalid email.", "/dashboard/account", "Go back")
		return
	}

	err := services.User.ChangeEmail(user.Id, email)
	if err != nil {
		slog.Error("change email", "err", err)

		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			// duplicated email
			renderDialog(w, "Error", "This email address is already used by another user.", "/dashboard/account", "Go back")
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	renderDialog(w, "Info", "An verification email has been sent to your new email address", "/dashboard/account", "Continue")
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
		renderDialog(w, "Error", "Unknown error", "/dashboard", "Go back")
		return
	}

	renderDialog(w, "Info", "Verification email sent", "/dashboard", "Continue")
}

func verifyEmailCallback(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		renderDialog(w, "Error", "Token is empty", "", "")
		return
	}

	err := services.User.VerifyEmail(token)
	if err != nil {
		slog.Error("verify email", "err", err)
		renderDialog(w, "Error", "Invalid token", "", "")
		return
	}

	renderDialog(w, "Info", "Your email is verified", "/dashboard", "Continue")
}

func changeEmailCallback(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		renderDialog(w, "Error", "Token is empty", "", "")
		return
	}

	err := services.User.ChangeEmailCallback(token)
	if err != nil {
		slog.Error("change email callback", "err", err)
		renderDialog(w, "Error", "Invalid token", "", "")
		return
	}

	renderDialog(w, "Info", "Your email is changed", "/dashboard", "Continue")
}

func myImages(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		page = 0
	}

	imageCount, err := services.Image.CountByUser(user.Id)
	if err != nil {
		slog.Error("my images", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	images, err := services.Image.FindByUser(user.Id, page)
	if err != nil {
		slog.Error("my images", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	render(w, "images", H{
		"user":       user,
		"images":     images,
		"page":       page,
		"total_page": int(math.Ceil(float64(imageCount) / 20)), // page size = 20
	})
}

func deleteImage(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	fileName := r.FormValue("file_name")

	if fileName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	img, err := services.Image.FindByFileName(fileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("delete image", "err", err)
		return
	}

	if img == nil {
		w.WriteHeader(http.StatusNotFound)
		renderDialog(w, "Error", "image not found", "/dashboard/images", "Go back")
		return
	}

	if !img.Uploader.Valid || img.Uploader.Int32 != int32(user.Id) {
		w.WriteHeader(http.StatusForbidden)
		renderDialog(w, "Error", "You do not have permission to delete this image", "/dashboard/images", "Go back")
		return
	}

	err = services.Image.Delete(img)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		renderDialog(w, "Error", "unknown error", "/dashboard/images", "Go back")
		slog.Error("delete image", "err", err)
		return
	}

	renderDialog(w, "Info", "Image deleted", "/dashboard/images", "Continue")

}
