package middleware

import (
	"html/template"
	"imgu2/services"
	"imgu2/templates"
	"log/slog"
	"net/http"
)

// ReCAPTCHA verifies "g-recaptcha-response" form value in request
//
// GET requests are ignored
func CAPTCHA(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// GET requests are not checked
			next.ServeHTTP(w, r)
			return
		}

		captcha, err := services.Setting.GetCAPTCHA()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("recaptcha middelware", "err", err)
			return
		}

		if captcha == services.CAPTCHA_NONE {
			// captcha disabled
			next.ServeHTTP(w, r)
			return
		}

		// recaptcha
		if captcha == services.CAPTCHA_RECAPTCHA {
			resp := r.FormValue("g-recaptcha-response")
			if resp == "" || !services.CAPTCHA.VerifyReCAPTCHA(resp) {
				w.WriteHeader(http.StatusUnauthorized)
				templates.Render(w, "dialog", map[string]any{
					"dialog": "Error",
					"msg":    "recpatcha verification failed",
					"link":   template.URL("javascript:history.back();"),
					"btn":    "Go Back",
				})
				return
			}
		}

		// hcaptcha
		if captcha == services.CAPTCHA_HCAPTCHA {
			resp := r.FormValue("h-captcha-response")
			if resp == "" || !services.CAPTCHA.VerifyHCaptcha(resp) {
				w.WriteHeader(http.StatusUnauthorized)
				templates.Render(w, "dialog", map[string]any{
					"dialog": "Error",
					"msg":    "hcpatcha verification failed",
					"link":   template.URL("javascript:history.back();"),
					"btn":    "Go Back",
				})
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
