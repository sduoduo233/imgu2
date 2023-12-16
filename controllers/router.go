package controllers

import (
	"imgu2/controllers/middleware"

	"github.com/go-chi/chi/v5"
	cmiddleware "github.com/go-chi/chi/v5/middleware"
)

func Route(r chi.Router) {
	r.Use(cmiddleware.Logger)
	r.Use(middleware.Auth)
	r.Use(middleware.CSRF)

	// auth
	r.Get("/login", login)
	r.With(middleware.ReCAPTCHA).Post("/login", doLogin)
	r.Get("/login/google", googleLogin)
	r.Get("/login/google/callback", googleLoginCallback)
	r.Get("/login/github", githubLogin)
	r.Get("/login/github/callback", githubLoginCallback)

	// register
	r.Get("/register", register)
	r.With(middleware.ReCAPTCHA).Post("/register", doRegister)

	// email callback
	r.Get("/verify-email", verifyEmailCallback)
	r.Get("/verify-email-change", changeEmailCallback)

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Get("/logout", logout)
	})

	// image
	r.Get("/i/{fileName}", downloadImage)
	r.Get("/preview/{fileName}", previewImage)

	// user dashboard
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Get("/dashboard", dashboardIndex)
		r.Get("/dashboard/account", accountSetting)
		r.With(middleware.ReCAPTCHA).Post("/dashboard/change-password", changePassword)
		r.With(middleware.ReCAPTCHA).Post("/dashboard/change-email", changeEmail)
		r.Post("/dashboard/change-username", changeUsername)
		r.Post("/dashboard/unlink", socialLoginUnlink)
		r.Get("/dashboard/verify-email", verifyEmail)
		r.With(middleware.ReCAPTCHA).Post("/dashboard/verify-email", doVerifyEmail)
		r.Get("/dashboard/images", myImages)
		r.Post("/dashboard/images/delete", deleteImage)
	})

	// admin dashboard
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Use(middleware.RequireAdmin)
		r.Get("/admin/settings", adminSettings)
		r.Post("/admin/settings", doAdminSettings)
		r.Get("/admin/storages", adminStorages)
		r.Get("/admin/storages/{id}", adminEditStorage)
		r.Post("/admin/storages/{id}", adminDoEditStorage)
		r.Post("/admin/storages/delete/{id}", adminStorageDelete)
		r.Post("/admin/storages", adminAddStorage)
	})

	r.Get("/", upload)
	r.With(middleware.ReCAPTCHA).Post("/upload", doUpload)
}
