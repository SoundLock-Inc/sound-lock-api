package service

import (
	"context"
	"errors"
	"sound_lock/internal/domain/models"
	"sound_lock/internal/service/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type profileServiceMocks struct {
	profileProviderMock *mocks.MockProfileProvider
}

// setUp — вспомогательная функция для инициализации ProfileService
// и связанных моков, используемых в тестах.
// Принимает контроллер gomock.Controller, возвращает сервис и структуру с моками.
func setUp(ctrl *gomock.Controller) (*ProfileService, *profileServiceMocks) {
	mocks := &profileServiceMocks{
		profileProviderMock: mocks.NewMockProfileProvider(ctrl),
	}

	profileService := newTestProfile(mocks.profileProviderMock)

	return profileService, mocks
}

// TestProfileService_ProfileUser_Success проверяет,
// что метод ProfileUser успешно возвращает данные пользователя,
// когда зависимость ProfileProvider возвращает корректные данные.
// Шаги:
//  1. Создаётся новый UUID пользователя.
//  2. Мок ProfileProvider настраивается на возврат UserDTO без ошибки.
//  3. Вызывается ProfileUser.
//  4. Проверяется, что ошибки нет и данные совпадают с ожидаемыми.
func TestProfileService_ProfileUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s, m := setUp(ctrl)

	userID := uuid.New()
	user := models.UserDTO{
		UUID:     userID,
		Email:    "email",
		Password: "password",
	}

	expectedUser := models.User{
		ID:    userID,
		Email: "email",
	}

	m.profileProviderMock.EXPECT().
		ReadUserByID(gomock.Any(), userID.String()).
		Return(user, nil)

	u, err := s.ProfileUser(context.Background(), userID)

	require.NoError(t, err)
	require.Equal(t, expectedUser, u)
	require.Equal(t, userID, u.ID)
	require.Equal(t, "email", u.Email)
}

// TestProfileService_ProfileUser_ErrorGetProfile проверяет,
// что метод ProfileUser корректно обрабатывает ошибку,
// возвращаемую ProfileProvider.
// Шаги:
//  1. Создаётся новый UUID пользователя.
//  2. Мок ProfileProvider настраивается на возврат ошибки.
//  3. Вызывается ProfileUser.
//  4. Проверяется, что вернулась ошибка и пустая модель пользователя.
func TestProfileService_ProfileUser_ErrorGetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	s, m := setUp(ctrl)

	userID := uuid.New()
	eUser := models.User{}

	m.profileProviderMock.EXPECT().
		ReadUserByID(gomock.Any(), userID.String()).
		Return(models.UserDTO{}, errors.New("Error reade user"))

	u, err := s.ProfileUser(context.Background(), userID)

	require.Error(t, err)
	require.Equal(t, eUser, u)
}
