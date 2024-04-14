package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"imgu2/controllers"
	"imgu2/db"
	"imgu2/libvips"
	"imgu2/services"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	listen := flag.String("listen", "127.0.0.1:3000", "listening address")
	debug := flag.Bool("debug", false, "debug logging")
	sqlitePath := flag.String("sqlite", "./db.sqlite", "path to sqlite database")
	flag.Parse()

	// logging
	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})))

	// load database
	db.Init(*sqlitePath)

	// libvips
	libvips.LibvipsInit()

	// initialize storage drivers
	err = services.Storage.Init()
	if err != nil {
		slog.Error("failed to initialize storage drivers", "err", err)
		panic(err)
	}

	// start scheduled tasks
	services.TaskStart()

	// initialize login providers
	err = services.Auth.InitOAuthProviders()
	if err != nil {
		slog.Error("failed to initialize oauth providers", "err", err)
		panic(err)
	}

	// http router
	r := chi.NewRouter()

	controllers.Route(r)

	slog.Info("server started", "listening", *listen)
	err = http.ListenAndServe(*listen, r)
	if err != nil {
		slog.Error("failed to listen and server", "err", err)
	}
}
