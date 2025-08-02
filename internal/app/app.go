package app

import (
	"context"
	"sound_lock/internal/config"
	"sound_lock/internal/handlers"
	"sound_lock/internal/http"
	"sound_lock/internal/lib/jwt"
	"sound_lock/internal/pkg"
	"sound_lock/internal/service"
	"sound_lock/internal/storage"
	"sound_lock/internal/storage/postgres"

	"github.com/labstack/echo/v4"
)

type App struct {
	*http.Server
}

func New(config *config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := pkg.SetUpLogger()

	log.Debug("sldsd")

	e := echo.New()

	srv := http.New(ctx, config.Port, config.Host, e)

	db := postgres.New(ctx, config)

	authStorage := storage.NewUser(db)
	refreshStorage := storage.NewRefreshToken(db)

	jwt := jwt.NewJWTManager(config.AccessTokenSigningKey, config.RefreshTokenSigningKey, config.AccessTokenTTL, config.RefreshTokenTTL)

	authService := service.NewAuth(
		authStorage,
		refreshStorage,
		jwt,
		config.AccessTokenTTL,
		config.RefreshTokenTTL,
	)

	h := handlers.New(e, authService)

	h.SetupHandlers()

	return &App{srv}
}
