package controllers

import (
	"fmt"
	"img2/services"
	"img2/templates"
	"io"
)

type H map[string]any

func render(w io.Writer, name string, data H) error {
	title, err := services.Setting.GetSiteName()
	if err != nil {
		return fmt.Errorf("get site name: %w", err)
	}
	data["title"] = title

	return templates.Render(w, name, data)
}

func renderError(w io.Writer, name string, err string) error {
	return render(w, name, H{
		"error": err,
	})
}

func renderDialog(w io.Writer, title, msg, link, btn string) error {
	return render(w, "dialog", H{
		"dialog": title,
		"msg":    msg,
		"link":   link,
		"btn":    btn,
	})
}
