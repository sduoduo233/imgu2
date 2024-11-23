package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	// http server
	server := &http.Server{Addr: *listen, Handler: r}
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen and serve", "err", err, "listen", *listen)
		}
	}()

	// graceful shutdown

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	sig := <-signalChan

	slog.Warn("shutdown", "sig", sig)

	services.TaskStop()

	err = server.Shutdown(context.Background())
	if err != nil {
		slog.Error("shutdown server", "err", err)
	}

}
