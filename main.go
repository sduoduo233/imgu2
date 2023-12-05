package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"imgu2/controllers"
	_ "imgu2/db"
	"imgu2/services"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	// initialize storage drivers
	err = services.Storage.Init()
	if err != nil {
		slog.Error("failed to initialize storage drivers", "err", err)
		panic(err)
	}

	// start scheduled tasks
	services.TaskStart()

	r := chi.NewRouter()

	controllers.Route(r)

	slog.Info("server started", "listening", ":3000")
	http.ListenAndServe(":3000", r)
}
