package app

import (
	"context"
	"sound_lock/internal/config"
	"sound_lock/internal/handlers"
	"sound_lock/internal/http"

	"github.com/labstack/echo/v4"
)

type App struct {
	*http.Server
}

func New(config *config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e := echo.New()

	srv := http.New(ctx, config.Port, config.Host, e)

	h := handlers.New(e)

	h.SetupHandlers()

	return &App{srv}
}
