package auth

import (
	"context"
	"net/http"
	"sound_lock/internal/config"
	"sound_lock/internal/domain/models"
	"sound_lock/internal/lib/jwt"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// AuthService определяет интерфейс для работы с аутентификацией пользователей.
type AuthService interface {
	LoginUser(ctx context.Context, email, password string) (models.Tokens, error)
	RegisterUser(ctx context.Context, email, password string) (string, error)
	LogoutUser(ctx context.Context, refreshToken string) error
	RefreshTokens(ctx context.Context, refreshToken string) (models.Tokens, error)
}

// Handlers — структура для хранения группы маршрутов Echo, связанных с аутентификацией.
type Handlers struct {
	authService AuthService // Service по работе с авторизацией
	cfg         *config.Config
}

// New — конструктор Handlers, принимает Echo-группу и возвращает объект Handlers.
func New(
	authService AuthService,
	cfg *config.Config,
) *Handlers {
	return &Handlers{
		authService: authService,
		cfg:         cfg,
	}
}

func newForTest(authService AuthService) *Handlers {
	return &Handlers{
		authService: authService,
		cfg:         &config.Config{},
	}
}

// SetupAuthHandlers — регистрирует эндпоинты, связанные с аутентификацией.
func (h *Handlers) SetupAuthHandlers(eg *echo.Group) {
	eg.POST("/login", h.login)       // Эндпоинт для входа в систему
	eg.POST("/register", h.register) // Эндпоинт для регистрации нового пользователя
	eg.Use(echojwt.WithConfig(jwt.GetConfig(h.cfg.AccessTokenSigningKey)))
	eg.PATCH("/refresh", h.refresh) // Эндпоинт для обновления токена
	eg.POST("/logout", h.logout)    // Эндпоинт для выхода пользователя из системы
}

// login — обработчик для входа пользователя в систему.
func (h *Handlers) login(c echo.Context) error {
	r := new(models.Regiseter)

	ctx := c.Request().Context()

	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": "title or body not valid",
		})
	}
	tokens, err := h.authService.LoginUser(ctx, r.Email, r.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, tokens)
}

// register — обработчик регистрации нового пользователя.
func (h *Handlers) register(c echo.Context) error {
	r := new(models.Regiseter)

	ctx := c.Request().Context()

	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
	}

	id, err := h.authService.RegisterUser(ctx, r.Email, r.Password)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"id": id,
	})
}

// logout — обработчик выхода пользователя из системы.
func (h *Handlers) logout(c echo.Context) error {
	r := new(models.RefreshRequest)

	ctx := c.Request().Context()

	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
	}

	err := h.authService.LogoutUser(ctx, r.Refresh)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
	}

	return c.NoContent(http.StatusOK)
}

// refresh — обработчик обновления JWT токена.
func (h *Handlers) refresh(c echo.Context) error {
	r := new(models.RefreshRequest)

	ctx := c.Request().Context()

	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
	}

	tokens, err := h.authService.RefreshTokens(ctx, r.Refresh)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, tokens)
}
