package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Salivare-DevHub/sso-auth-server/internal/app"
	"github.com/Salivare-DevHub/sso-auth-server/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	application, err := app.New(log, cfg)
	if err != nil {
		log.Error("failed to init application", slog.String("err", err.Error()))
		os.Exit(1)
	}

	go application.RPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	log.Info("Shutting down...")

	application.RPCSrv.Stop()

	log.Info("Goodbye!")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case config.EnvLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case config.EnvDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case config.EnvProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
