package profile

import (
	"context"
	"net/http"
	"sound_lock/internal/domain/models"
	"sound_lock/internal/lib/jwt"

	j "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Service определяет интерфейс сервиса для работы с профилем пользователя.
// Реализация должна предоставлять методы для получения и обновления данных профиля.
type Service interface {
	ProfileUser(ctx context.Context, userID uuid.UUID) (models.User, error)
	ChangeProfileUser(ctx context.Context, userDTO models.UserDTO) (models.User, error)
}

// Handlers содержит обработчики HTTP-запросов для работы с профилем.
// Использует ProfileService для доступа к бизнес-логике.
type Handlers struct {
	profileService Service
}

// New создает новый экземпляр Handlers с заданным сервисом профиля.
func New(profileService Service) *Handlers {
	return &Handlers{
		profileService: profileService,
	}
}

// SetupProfileHandler регистрирует обработчики для работы с профилем в Echo Group.
func (h *Handlers) SetupProfileHandler(eg *echo.Group) {
	eg.GET("", h.getProfile)
	eg.PATCH("", h.updateProfile)
}

// getProfile обрабатывает запрос на получение данных профиля пользователя.
func (h *Handlers) getProfile(c echo.Context) error {
	ctx := c.Request().Context()

	u := c.Get("user").(*j.Token)
	claims, ok := u.Claims.(*jwt.TokenClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]any{
			"error": "Unauthorized",
		})
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Unauthorized",
		})
	}

	user, err := h.profileService.ProfileUser(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, user)
}

// updateProfile обрабатывает запрос на обновление данных профиля пользователя.
func (h *Handlers) updateProfile(c echo.Context) error {
	return nil
}
