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
	funcs := template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
		"minus": func(a int, b int) int {
			return a - b
		},
		"loop": func(a int, b int) []int {
			arr := make([]int, b-a)
			for i := 0; i < b-a; i++ {
				arr[i] = a + i
			}
			return arr
		},
	}

	var err error
	t, err = template.New("").Funcs(funcs).ParseFS(fs, "html/*.html")
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
