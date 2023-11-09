package controllers

import (
	"img2/services"
	"img2/templates"
	"io"
	"log/slog"
)

type H map[string]any

func render(w io.Writer, name string, data H) {
	title, err := services.Setting.GetSiteName()
	if err != nil {
		slog.Error("render template: site name", "err", err)
		return
	}
	data["title"] = title

	err = templates.Render(w, name, data)
	if err != nil {
		slog.Error("render template", "err", err, "name", name, "data", data)
		return
	}
}

func renderError(w io.Writer, name string, err string) {
	render(w, name, H{
		"error": err,
	})
}

func renderDialog(w io.Writer, title, msg, link, btn string) {
	render(w, "dialog", H{
		"dialog": title,
		"msg":    msg,
		"link":   link,
		"btn":    btn,
	})
}
