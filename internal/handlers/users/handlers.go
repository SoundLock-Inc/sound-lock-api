package users

import (
	"github.com/labstack/echo/v4"
)

// Handlers — структура для роутов, относящихся к пользователю.
// Содержит группу маршрутов Echo, привязанных к /api/v1/users.
type Handlers struct {
	eg *echo.Group // Echo-группа с базовым путём /users
}

// New — конструктор для Handlers, принимает группу маршрутов Echo
func New(eg *echo.Group) *Handlers {
	return &Handlers{eg: eg}
}

// SetupAudioHandlers — регистрирует эндпоинты, связанные с пользователем.
func (h *Handlers) SetupUsersHandlers() {
	h.eg.GET("/:id", h.getUser)       // Получить пользователя по ID
	h.eg.GET("", h.getUsers)          // Получить список всех пользователей
	h.eg.PATCH("/:id", h.updateUser)  // Обновить данные пользователя по ID
	h.eg.DELETE("/:id", h.deleteUser) // Удалить пользователя по ID
}

// getUser — обработчик запроса на получение информации об одном пользователе
func (h *Handlers) getUser(c echo.Context) error {
	return nil
}

// getUsers — обработчик запроса на получение списка пользователей
func (h *Handlers) getUsers(c echo.Context) error {
	return nil
}

// updateUser — обработчик запроса на обновление информации пользователя
func (h *Handlers) updateUser(c echo.Context) error {
	return nil
}

// deleteUser — обработчик запроса на удаление пользователя
func (h *Handlers) deleteUser(c echo.Context) error {
	return nil
}
