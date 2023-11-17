package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"img2/controllers"
	_ "img2/db"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	r := chi.NewRouter()

	controllers.Route(r)

	slog.Info("server started", "listening", ":3000")
	http.ListenAndServe(":3000", r)
}
