package templates

import (
	"embed"
	"fmt"
	"html/template"
	"imgu2/i18n"
	"io"
	"log/slog"
	"net/url"
	"time"
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
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				panic("dict: invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					panic("dict: keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict
		},
		"timestamp": func(t time.Time) int64 {
			return t.Unix()
		},
		"addParameter": func(u, name string, value any) string {
			// add a new parameter to an existing url

			s, err := url.ParseRequestURI(u)
			if err != nil {
				panic(u + ", " + err.Error())
			}

			q := s.Query()
			q.Set(name, fmt.Sprintf("%v", value))
			s.RawQuery = q.Encode()

			return s.RequestURI()
		},
		"tr": func(key string) string {
			return i18n.T(key)
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
