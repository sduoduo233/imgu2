package controllers

import (
	"encoding/json"
	"imgu2/i18n"
	"imgu2/services"
	"imgu2/templates"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type H map[string]any

func render(w io.Writer, name string, data H) {
	title, err := services.Setting.GetSiteName()
	if err != nil {
		slog.Error("render template: site name", "err", err)
		return
	}
	data["title"] = title

	captcha, err := services.Setting.GetCAPTCHA()
	if err != nil {
		slog.Error("render template: captcha", "err", err)
		return
	}

	if captcha == services.CAPTCHA_RECAPTCHA {
		// recaptcha site key
		recaptcha, err := services.Setting.GetReCaptchaClient()
		if err != nil {
			slog.Error("render template: captcha client", "err", err)
			return
		}

		data["recaptcha_client"] = recaptcha
	}

	if captcha == services.CAPTCHA_HCAPTCHA {
		// hcaptcha site key
		hcaptcha, err := services.Setting.GetHCaptchaClient()
		if err != nil {
			slog.Error("render template: captcha client", "err", err)
			return
		}

		data["hcaptcha_client"] = hcaptcha
	}

	err = templates.Render(w, name, data)
	if err != nil {
		slog.Error("render template", "err", err, "name", name, "data", data)
		return
	}
}

func renderDialog(w io.Writer, title, msg string, link any, btn string) {
	render(w, "dialog", H{
		"dialog": title,
		"msg":    msg,
		"link":   link,
		"btn":    btn,
	})
}

func writeJSON(w io.Writer, m H) {
	b, err := json.Marshal(m)
	if err != nil {
		slog.Error("write json", "err", err)
		return
	}

	_, err = io.WriteString(w, string(b))
	if err != nil {
		slog.Error("write json", "err", err)
	}
}

func setCookie(w http.ResponseWriter, name string, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 30),
	})
}

// generate a new csrf token and add it to cookies
func csrfToken(w http.ResponseWriter) string {
	t := services.RandomHexString(8)
	setCookie(w, "CSRF_TOKEN", t)
	return t
}

// shortcut for i18n.T
func tr(key string) string {
	return i18n.T(key)
}
