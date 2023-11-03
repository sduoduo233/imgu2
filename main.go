package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"img2/controllers"
	_ "img2/db"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	r := chi.NewRouter()

	controllers.Route(r)

	slog.Info("server started", "listening", ":3000")
	http.ListenAndServe(":3000", r)
}
