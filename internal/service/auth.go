package service

import (
	"context"
	"fmt"
	"sound_lock/internal/domain/models"
	"sound_lock/internal/lib/jwt"
	"sound_lock/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type JWTMethods interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ParseAccessToken(tokenString string) (*jwt.TokenClaims, error)
	ParseRefreshToken(tokenString string) (*jwt.TokenClaims, error)
}

// UserSaver определяет интерфейс для сохранения данных о пользователях.
type UserSaver interface {
	CreateUser(ctx context.Context, email, password string) (string, error)
}

// UserProvider определяет интерфейс для получения данных о пользователях.
type UserProvider interface {
	ReadUser(ctx context.Context, email string) (models.UserDTO, error)
}

// RefreshTokenSaver определяет интерфейс для создания refresh-токенов.
type RefreshTokenSaver interface {
	CreateRefreshToken(ctx context.Context, userID string, token string, expiresAt time.Time) error
}

// RefreshTokenProvider определяет интерфейс для обновления refresh-токенов.
type RefreshTokenProvider interface {
	UpdateRefreshToken(ctx context.Context, userID string, token, oldToken string, expiresAt time.Time) error
}

// RefreshTokenRemover определяет интерфейс для удаления refresh-токенов.
type RefreshTokenRemover interface {
	DeleteRefreshToken(ctx context.Context, userID, token string) error
}

// Auth реализует бизнес-логику, связанную с аутентификацией и управлением пользователями.
type Auth struct {
	userSaver            UserSaver
	userProvider         UserProvider
	refreshTokenSaver    RefreshTokenSaver
	refreshTokenProvider RefreshTokenProvider
	refreshTokenRemover  RefreshTokenRemover
	jwt                  JWTMethods
	accessTokenTTL       time.Duration
	refreshTokenTTL      time.Duration
}

func newAuthTest(
	userProvider UserProvider,
	userSaver UserSaver,
	refreshTokenSaver RefreshTokenSaver,
	refreshTokenProvider RefreshTokenProvider,
	refreshTokenRemover RefreshTokenRemover,
	jwtMethods JWTMethods,
	accessTokenTTL, refreshTokenTTL time.Duration,
) *Auth {
	return &Auth{
		userSaver:            userSaver,
		userProvider:         userProvider,
		refreshTokenSaver:    refreshTokenSaver,
		refreshTokenProvider: refreshTokenProvider,
		refreshTokenRemover:  refreshTokenRemover,
		jwt:                  jwtMethods,
		accessTokenTTL:       accessTokenTTL,
		refreshTokenTTL:      refreshTokenTTL,
	}
}

// NewAuth создает новый экземпляр Auth с переданными зависимостями.
func NewAuth(
	userStorage *storage.User,
	refreshTokenStorage *storage.RefreshTokenStorage,
	jwt *jwt.JWTManager,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Auth {
	return &Auth{
		jwt:                  jwt,
		userSaver:            userStorage,
		userProvider:         userStorage,
		refreshTokenSaver:    refreshTokenStorage,
		refreshTokenProvider: refreshTokenStorage,
		refreshTokenRemover:  refreshTokenStorage,
		accessTokenTTL:       accessTokenTTL,
		refreshTokenTTL:      refreshTokenTTL,
	}
}

// RegisterUser регистрирует нового пользователя, хэшируя пароль и сохраняя данные в хранилище.
// Возвращает UUID созданного пользователя.
func (a *Auth) RegisterUser(ctx context.Context, email, password string) (string, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	uuid, err := a.userSaver.CreateUser(ctx, email, string(passHash))
	if err != nil {
		return "", err
	}

	return uuid, nil
}

// LoginUser выполняет аутентификацию пользователя, сравнивает пароль,
// генерирует access и refresh токены и возвращает их.
func (a *Auth) LoginUser(ctx context.Context, email, password string) (models.Tokens, error) {
	user, err := a.userProvider.ReadUser(ctx, email)
	if err != nil {
		return models.Tokens{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return models.Tokens{}, err
	}

	tokens := models.Tokens{}

	access, err := a.jwt.GenerateAccessToken(user.UUID.String())
	if err != nil {
		return models.Tokens{}, err
	}
	tokens.AccessToken = access

	refresh, err := a.jwt.GenerateRefreshToken(user.UUID.String())
	if err != nil {
		return models.Tokens{}, err
	}
	tokens.RefreshToken = refresh

	err = a.refreshTokenSaver.CreateRefreshToken(ctx, user.UUID.String(), refresh, time.Now().Add(a.refreshTokenTTL))
	if err != nil {
		return models.Tokens{}, err
	}

	return tokens, nil
}

// LogoutUser выполняет выход пользователя.
func (a *Auth) LogoutUser(ctx context.Context, refreshToken string) error {
	claims, err := a.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("error parse token %w", err)
	}

	err = a.refreshTokenRemover.DeleteRefreshToken(ctx, claims.UserID, refreshToken)
	if err != nil {
		return fmt.Errorf("failed remove token %w", err)
	}

	return nil
}

// RefreshTokens обновляет пару access/refresh токенов по предоставленному refresh токену.
func (a *Auth) RefreshTokens(ctx context.Context, refreshToken string) (models.Tokens, error) {
	var tokens models.Tokens

	claims, err := a.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return models.Tokens{}, fmt.Errorf("error parse token %w", err)
	}

	access, err := a.jwt.GenerateAccessToken(claims.UserID)
	if err != nil {
		return models.Tokens{}, fmt.Errorf("failed generate access token %w", err)
	}
	tokens.AccessToken = access

	refresh, err := a.jwt.GenerateRefreshToken(claims.UserID)
	if err != nil {
		return models.Tokens{}, fmt.Errorf("failed generate refresh token %w", err)
	}
	tokens.RefreshToken = refresh

	err = a.refreshTokenProvider.UpdateRefreshToken(ctx, claims.UserID, refresh, refreshToken, time.Now().Add(a.refreshTokenTTL))
	if err != nil {
		return models.Tokens{}, fmt.Errorf("failed update refresh token %w", err)
	}

	return tokens, nil
}
