package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sound_lock/internal/app"
	"sound_lock/internal/config"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()

	app := app.New(cfg)

	go app.MustRun()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	slog.Info("shutdown initiated...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	app.Stop(shutdownCtx)
	slog.Info("server stopped...")

}
