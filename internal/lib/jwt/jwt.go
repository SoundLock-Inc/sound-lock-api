package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// JWTManager отвечает за генерацию и валидацию access/refresh токенов
type JWTManager struct {
	accessSecret    []byte        // секретный ключ для подписания access-токенов
	refreshSecret   []byte        // секретный ключ для подписания refresh-токенов
	accessTokenTTL  time.Duration // время жизни access-токена
	refreshTokenTTL time.Duration // время жизни refresh-токена
}

// NewJWTManager конструктор JWTManager, принимает секреты и TTL
func NewJWTManager(
	accessSecret, refreshSecret string,
	accessTTL, refreshTTL time.Duration,
) *JWTManager {
	return &JWTManager{
		accessSecret:    []byte(accessSecret),  // преобразуем строку в []byte для HMAC
		refreshSecret:   []byte(refreshSecret), // отдельный ключ для refresh токена
		accessTokenTTL:  accessTTL,             // время жизни access токена (обычно 15 мин)
		refreshTokenTTL: refreshTTL,            // время жизни refresh токена (обычно дни/недели)
	}
}

// TokenClaims описывает полезную нагрузку токена
// Включает встроенные поля RegisteredClaims (iss, exp, iat и т.д.) + UserID
type TokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

func GetConfig(accessSecret string) echojwt.Config {
	var jwtConfig = echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(TokenClaims)
		},
		SigningKey: []byte(accessSecret),
	}
	return jwtConfig
}

// GenerateAccessToken создает новый access-токен для конкретного пользователя
func (m *JWTManager) GenerateAccessToken(userID string) (string, error) {
	now := time.Now()
	// Claims (данные в теле токена)
	accessClaims := &TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenTTL)), // срок действия
			IssuedAt:  jwt.NewNumericDate(now),                       // время выпуска
		},
	}
	// Создаем токен с алгоритмом HMAC-SHA256 и подписываем секретом
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(m.accessSecret)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// GenerateRefreshToken создает refresh-токен для обновления access-токена
func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()

	refreshClaims := &TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenTTL)), // длинный срок действия
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	// Подписываем refresh-токен отдельным секретом (рекомендуется отличать от access)
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(m.refreshSecret)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

// ParseAccessToken проверяет и извлекает данные из строки access токена.
// Возвращает структуру с данными токена (claims) или ошибку.
func (m *JWTManager) ParseAccessToken(tokenString string) (*TokenClaims, error) {
	return m.parse(tokenString, m.accessSecret)
}

// ParseRefreshToken проверяет и извлекает данные из строки refresh токена.
// Возвращает структуру с данными токена (claims) или ошибку.
func (m *JWTManager) ParseRefreshToken(tokenString string) (*TokenClaims, error) {
	return m.parse(tokenString, m.refreshSecret)
}

// parse выполняет базовую операцию верификации JWT токена и извлечения claims.
// Возвращает:
// - TokenClaims при успешной проверке
// - Ошибку при любой проблеме верификации или несоответствии структуры
func (m *JWTManager) parse(tokenString string, secret []byte) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&TokenClaims{},
		func(token *jwt.Token) (any, error) {
			return secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil

}
