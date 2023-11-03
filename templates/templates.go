package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log/slog"
)

//go:embed html/*
var fs embed.FS

var t *template.Template

func init() {
	var err error
	t, err = template.ParseFS(fs, "html/*.html")
	if err != nil {
		slog.Error("parse templates", "err", err)
		panic(err)
	}
}

func Render(writer io.Writer, name string, data any) error {
	err := t.ExecuteTemplate(writer, name+".html", data)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	return nil
}
