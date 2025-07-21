package handlers

import (
	audio "sound_lock/internal/handlers/audio_password"
	"sound_lock/internal/handlers/auth"
	"sound_lock/internal/handlers/users"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Handlers — структура, содержащая экземпляр Echo,
// чтобы централизованно настраивать все роуты приложения.
type Handlers struct {
	e *echo.Echo
}

// New — конструктор, создающий новый экземпляр Handlers
func New(e *echo.Echo) *Handlers {
	return &Handlers{e: e}
}

// SetupHandlers — метод, настраивающий все маршруты и middleware приложения.
func (h *Handlers) SetupHandlers() {
	// Восстанавливает приложение после паники и логирует ошибки
	h.e.Use(middleware.Recover())

	// Включает логгирование всех HTTP-запросов
	h.e.Use(middleware.Logger())

	// Группа всех API-эндпоинтов с префиксом /api/v1
	api := h.e.Group("/api/v1")

	// Группа эндпоинтов для аутентификации (/api/v1/auth)
	authGroup := api.Group("/auth")
	auth.New(authGroup).SetupAuthHandlers()

	// Группа эндпоинтов, связанных с пользователями (/api/v1/users)
	usersGroup := api.Group("/users")
	users.New(usersGroup).SetupUsersHandlers()

	// Группа эндпоинтов для аудио-паролей (/api/v1/audio-password)
	audioGroup := api.Group("/audio-password")
	audio.New(audioGroup).SetupAudioHandlers()
}
