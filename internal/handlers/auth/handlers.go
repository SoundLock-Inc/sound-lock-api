package auth

import (
	"github.com/labstack/echo/v4"
)

// Handlers — структура для хранения группы маршрутов Echo, связанных с аутентификацией.
type Handlers struct {
	eg *echo.Group // Echo-группа для /auth
}

// New — конструктор Handlers, принимает Echo-группу и возвращает объект Handlers.
func New(eg *echo.Group) *Handlers {
	return &Handlers{eg: eg}
}

// SetupAuthHandlers — регистрирует эндпоинты, связанные с аутентификацией.
func (h *Handlers) SetupAuthHandlers() {
	h.eg.POST("/login", h.login)       // Эндпоинт для входа в систему
	h.eg.POST("/register", h.register) // Эндпоинт для регистрации нового пользователя
	h.eg.POST("/refresh", h.refresh)   // Эндпоинт для обновления токена
}

// login — обработчик для входа пользователя в систему.
func (h *Handlers) login(c echo.Context) error {
	return nil
}

// register — обработчик регистрации нового пользователя.
func (h *Handlers) register(c echo.Context) error {
	return nil
}

// refresh — обработчик обновления JWT токена.
func (h *Handlers) refresh(c echo.Context) error {
	return nil
}
