package controllers

import (
	"encoding/json"
	"imgu2/services"
	"imgu2/templates"
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

func renderDialog(w io.Writer, title, msg, link, btn string) {
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
