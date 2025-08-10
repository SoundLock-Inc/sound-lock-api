package profile

import (
	"context"
	"net/http"
	"sound_lock/internal/domain/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ProfileService interface {
	ProfileUser(ctx context.Context, userID uuid.UUID) (models.User, error)
}

type Handlers struct {
	profileService ProfileService
}

func New(profileService ProfileService) *Handlers {
	return &Handlers{
		profileService: profileService,
	}
}

func (h *Handlers) SetupProfileHandler(eg *echo.Group) {
	eg.GET("", h.profile)
}

func (h *Handlers) profile(c echo.Context) error {
	ctx := c.Request().Context()
	// Убрать!!!!
	userID := uuid.New()

	user, err := h.profileService.ProfileUser(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, user)
}
