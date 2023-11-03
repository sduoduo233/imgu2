package middleware

import (
	"context"
	"img2/db"
	"img2/services"
	"io"
	"log/slog"
	"net/http"
)

// add user to the context if the cookie is valid
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("TOKEN")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		token := cookie.Value

		user, err := services.Session.Find(token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("auth: find session", "err", err)
			return
		}

		if user == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "USER", user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// abort the request if user is not present in request context
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value("USER").(*db.User)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("USER").(*db.User)
		if !ok {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "403 Forbidden")
			return
		}

		if user.Role != services.RoleAdmin {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "403 Forbidden")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// panic if user is nil
func MustGetUser(ctx context.Context) *db.User {
	user := ctx.Value("USER").(*db.User)
	return user
}

func GetUser(ctx context.Context) *db.User {
	user, ok := ctx.Value("USER").(*db.User)
	if !ok {
		return nil
	}
	return user
}
