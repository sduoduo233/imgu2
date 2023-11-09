package controllers

import (
	"img2/controllers/middleware"

	"github.com/go-chi/chi/v5"
	cmiddleware "github.com/go-chi/chi/v5/middleware"
)

func Route(r chi.Router) {
	r.Use(cmiddleware.Logger)
	r.Use(middleware.Auth)

	r.Get("/login", login)
	r.Post("/login", doLogin)
	r.Get("/login/google", googleLogin)
	r.Get("/login/google/callback", googleLoginCallback)
	r.Get("/login/github", githubLogin)
	r.Get("/login/github/callback", githubLoginCallback)

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
	})

	// admin dashboard
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Use(middleware.RequireAdmin)
		r.Get("/admin", adminIndex)
		r.Get("/admin/settings", adminSettings)
		r.Post("/admin/settings", doAdminSettings)
	})
}
