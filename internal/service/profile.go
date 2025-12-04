package service

import (
	"context"
	"sound_lock/internal/domain/models"
	"sound_lock/internal/storage"

	"github.com/google/uuid"
)

// ProfileProvider — интерфейс, который должен реализовывать
// любой источник данных, способный вернуть информацию о пользователе
// по его ID. Например, может быть реализован через слой хранения (storage.User)
// или через внешний сервис.
type ProfileProvider interface {
	ReadUserByID(ctx context.Context, userID string) (models.UserDTO, error)
	UpdateUser(ctx context.Context, userDTO models.UserDTO) (models.UserDTO, error)
}

// ProfileService — слой бизнес-логики, отвечающий за операции с профилем пользователя.
type ProfileService struct {
	profileProvider ProfileProvider
}

// newTestProfile — вспомогательный конструктор для тестов,
// который позволяет подставить кастомную реализацию ProfileProvider
// (например, мок).
func newTestProfile(profileProvider ProfileProvider) *ProfileService {
	return &ProfileService{
		profileProvider: profileProvider,
	}
}

// NewProfile — основной конструктор сервиса,
// принимает указатель на storage.User как реализацию ProfileProvider.
func NewProfile(userStorage *storage.User) *ProfileService {
	return &ProfileService{
		profileProvider: userStorage,
	}
}

// ProfileUser — возвращает публичную модель пользователя (models.User)
// по его UUID. Выполняет:
//  1. Чтение DTO пользователя через profileProvider.
//  2. Преобразование DTO в публичную модель.
//
// Если в процессе чтения возникнет ошибка, возвращает пустую модель и ошибку.
func (p *ProfileService) ProfileUser(ctx context.Context, userID uuid.UUID) (models.User, error) {
	var user models.User

	userDTO, err := p.profileProvider.ReadUserByID(ctx, userID.String())
	if err != nil {
		return models.User{}, err
	}

	userDTO.ToUser(&user)

	return user, nil
}

func (p *ProfileService) ChangeProfileUser(ctx context.Context, userDTO models.UserDTO) (models.User, error) {
	var user models.User

	u, err := p.profileProvider.UpdateUser(ctx, userDTO)
	if err != nil {
		return models.User{}, err
	}

	u.ToUser(&user)

	return user, nil
}
