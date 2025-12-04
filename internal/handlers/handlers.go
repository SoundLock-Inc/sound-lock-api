package handlers

import (
	"sound_lock/internal/config"
	audio "sound_lock/internal/handlers/audio_password"
	"sound_lock/internal/handlers/auth"
	"sound_lock/internal/handlers/profile"
	"sound_lock/internal/handlers/users"
	"sound_lock/internal/lib/jwt"
	"sound_lock/internal/service"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Handlers — структура, содержащая экземпляр Echo,
// чтобы централизованно настраивать все роуты приложения.
type Handlers struct {
	e              *echo.Echo
	cfg            *config.Config
	authService    *service.Auth
	profileService *service.ProfileService
}

// New — конструктор, создающий новый экземпляр Handlers
func New(
	e *echo.Echo,
	authService *service.Auth,
	profileService *service.ProfileService,
	cfg *config.Config,
) *Handlers {
	return &Handlers{
		e:              e,
		authService:    authService,
		profileService: profileService,
		cfg:            cfg,
	}
}

// SetupHandlers — метод, настраивающий все маршруты и middleware приложения.
func (h *Handlers) SetupHandlers(accessSecret string) {
	// Восстанавливает приложение после паники и логирует ошибки
	h.e.Use(middleware.Recover())

	// Включает логгирование всех HTTP-запросов
	h.e.Use(middleware.Logger())

	// Группа всех API-эндпоинтов с префиксом /api/v1
	api := h.e.Group("/api/v1")

	// Группа эндпоинтов для аутентификации (/api/v1/auth)
	authGroup := api.Group("/auth")
	auth.New(h.authService, h.cfg).SetupAuthHandlers(authGroup)

	// Группа эндпоинтов для профиля (/api/v1/profile)
	profileGroup := api.Group("/profile")
	profileGroup.Use(echojwt.WithConfig(jwt.GetConfig(h.cfg.AccessTokenSigningKey)))
	profile.New(h.profileService).SetupProfileHandler(profileGroup)

	// Группа эндпоинтов, связанных с пользователями (/api/v1/users)
	usersGroup := api.Group("/users")
	usersGroup.Use(echojwt.WithConfig(jwt.GetConfig(h.cfg.AccessTokenSigningKey)))
	users.New(usersGroup).SetupUsersHandlers()

	// Группа эндпоинтов для аудио-паролей (/api/v1/audio-password)
	audioGroup := api.Group("/audio-password")
	audioGroup.Use(echojwt.WithConfig(jwt.GetConfig(h.cfg.AccessTokenSigningKey)))
	audio.New(audioGroup).SetupAudioHandlers()
}
