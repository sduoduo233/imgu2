package controllers

import (
	"imgu2/controllers/middleware"
	"net/http"
)

func adminIndex(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r.Context())

	render(w, "admin", H{
		"user": user,
	})
}
