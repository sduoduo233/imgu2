package controllers

import (
	"img2/controllers/middleware"

	"github.com/go-chi/chi/v5"
	cmiddleware "github.com/go-chi/chi/v5/middleware"
)

func Route(r chi.Router) {
	r.Use(cmiddleware.Logger)
	r.Use(middleware.Auth)

	// auth
	r.Get("/login", login)
	r.Post("/login", doLogin)
	r.Get("/login/google", googleLogin)
	r.Get("/login/google/callback", googleLoginCallback)
	r.Get("/login/github", githubLogin)
	r.Get("/login/github/callback", githubLoginCallback)

	r.Get("/verify-email", verifyEmailCallback)

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Get("/logout", logout)
	})

	// user dashboard
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Get("/dashboard", dashboardIndex)
		r.Get("/dashboard/account", accountSetting)
		r.Post("/dashboard/change-password", changePassword)
		r.Post("/dashboard/change-email", changeEmail)
		r.Post("/dashboard/change-username", changeUsername)
		r.Post("/dashboard/unlink", socialLoginUnlink)
		r.Get("/dashboard/verify-email", verifyEmail)
		r.Post("/dashboard/verify-email", doVerifyEmail)
	})

	// admin dashboard
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Use(middleware.RequireAdmin)
		r.Get("/admin", adminIndex)
		r.Get("/admin/settings", adminSettings)
		r.Post("/admin/settings", doAdminSettings)
		r.Get("/admin/storages", adminStorages)
		r.Get("/admin/storages/{id}", adminEditStorage)
		r.Post("/admin/storages/{id}", adminDoEditStorage)
		r.Post("/admin/storages/delete/{id}", adminStorageDelete)
		r.Post("/admin/storages", adminAddStorage)
	})

	r.Get("/", upload)
	r.Post("/upload", doUpload)
}
