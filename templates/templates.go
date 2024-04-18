package templates

import (
	"embed"
	"fmt"
	"html/template"
	"imgu2/i18n"
	"io"
	"log/slog"
	"math"
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
		"formatFileSize": func(n int) string {
			// add human readable units to number of bytes

			symbols := []string{"kB", "MB", "GB"}

			if n < 1000 {
				return fmt.Sprintf("%d B", n)
			}

			for i, symbol := range symbols {
				if n >= int(math.Pow10((i+1)*3)) && n < int(math.Pow10((i+2)*3)) {

					return fmt.Sprintf("%.1f %s", float64(n)/math.Pow10((i+1)*3), symbol)
				}
			}

			return fmt.Sprintf("%.1f %s", float64(n)/math.Pow10(6), symbols[2])
		},
		"formatDate": func(t time.Time) string {
			// return a date string accepted by HTML input tag
			// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input/date

			return t.UTC().Format("2006-01-02")
		},
		"formatTime": func(t time.Time) string {
			// return a time string accepted by HTML input tag
			// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input/time

			return t.UTC().Format("15:04")
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
