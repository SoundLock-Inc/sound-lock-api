package profile

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sound_lock/internal/domain/models"
	"sound_lock/internal/handlers/profile/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

// TestHandlers_Profile_Success проверяет корректную работу HTTP-хэндлера профиля,
// когда сервис успешно возвращает данные пользователя.
// Сценарий:
//  1. Создаем мок сервиса профиля (ProfileService).
//  2. Настраиваем хэндлер и роутинг.
//  3. Имитируем HTTP-запрос GET /profile.
//  4. Мок возвращает корректную структуру models.User без ошибки.
//  5. Проверяем, что статус ответа — 200 OK и тело ответа корректно сериализуется в User.
func TestHandlers_Profile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProfileService := mocks.NewMockProfileService(ctrl)

	h := New(mockProfileService)
	e := echo.New()

	group := e.Group("/profile")
	h.SetupProfileHandler(group)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	userID := uuid.New()
	user := models.User{Email: "email", ID: userID}

	mockProfileService.EXPECT().ProfileUser(gomock.Any(), gomock.Any()).Return(user, nil)

	h.profile(c)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &user))
}

// TestHandlers_Profile_ErrorGetProfile проверяет поведение хэндлера профиля,
// если сервис возвращает ошибку при получении данных пользователя.
// Сценарий:
//  1. Создаем мок сервиса профиля.
//  2. Настраиваем хэндлер и роутинг.
//  3. Имитируем HTTP-запрос GET /profile.
//  4. Мок возвращает пустого пользователя и ошибку.
//  5. Проверяем, что хэндлер возвращает статус 500 Internal Server Error.
func TestHandlers_Profile_ErrorGetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProfileService := mocks.NewMockProfileService(ctrl)

	h := New(mockProfileService)
	e := echo.New()

	group := e.Group("/profile")
	h.SetupProfileHandler(group)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	user := models.User{}

	mockProfileService.EXPECT().ProfileUser(gomock.Any(), gomock.Any()).Return(user, errors.New("Error get profile"))

	h.profile(c)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
