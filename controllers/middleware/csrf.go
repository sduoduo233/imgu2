package middleware

import (
	"io"
	"net/http"
)

// CSRF verifies the CSRF token in request body and cookie on every POST request.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// GET requests are not checked
			next.ServeHTTP(w, r)
			return
		}

		token := r.FormValue("csrf_token")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "csrf check failed")
			return
		}

		cookie, err := r.Cookie("CSRF_TOKEN")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "csrf check failed")
			return
		}

		if token != cookie.Value {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "csrf check failed")
			return
		}

		next.ServeHTTP(w, r)
	})
}
