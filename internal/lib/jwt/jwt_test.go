package jwt_test

import (
	"sound_lock/internal/lib/jwt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestJWT_GenerateAccessToken_Success проверяет успешную генерацию access токена.
// Убеждается, что:
// - Токен создается без ошибок
// - Результирующая строка токена не пустая
func TestJWT_GenerateAccessToken_Success(t *testing.T) {
	m := jwt.NewJWTManager("access-secret", "refresh-secret", time.Minute, time.Hour)
	tokenString, err := m.GenerateAccessToken("user-123")
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)
}

// TestJWT_ParseAccessToken_Success проверяет корректность парсинга валидного access токена.
// Тестирует:
// - Успешное извлечение claims из сгенерированного токена
// - Соответствие идентификатора пользователя
// - Корректность срока действия токена (с допуском ±2 секунды)
func TestJWT_ParseAccessToken_Success(t *testing.T) {
	m := jwt.NewJWTManager("access-secret", "refresh-secret", time.Minute, time.Hour)

	tokenString, err := m.GenerateAccessToken("user-123")

	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	claims, err := m.ParseAccessToken(tokenString)

	require.NoError(t, err)
	require.Equal(t, "user-123", claims.UserID)
	require.WithinDuration(t, time.Now().Add(time.Minute), claims.ExpiresAt.Time, time.Second*2)
}

// TestJWT_ParseAccessToken_Error проверяет обработку невалидных access токенов.
// Сценарии:
// - Парсинг заведомо поврежденной строки токена
// - Парсинг токена, подписанного неверным секретом
// Ожидается:
// - Возврат ошибки во всех проблемных случаях
func TestJWT_ParseAccessToken_Error(t *testing.T) {
	m := jwt.NewJWTManager("access-secret", "refresh-secret", time.Minute, time.Hour)

	_, err := m.ParseAccessToken("invalid.token.string")

	require.Error(t, err)

	tokenString, err := m.GenerateAccessToken("user-123")

	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	badM := jwt.NewJWTManager("bad-secret", "refresh-secret", time.Minute, time.Hour)

	_, err = badM.ParseAccessToken(tokenString)
	require.Error(t, err)

}

// TestJWT_GenerateRefreshToken_Success проверяет успешную генерацию refresh токена.
// Убеждается, что:
// - Токен создается без ошибок
// - Результирующая строка токена не пустая
func TestJWT_GenerateRefreshToken_Success(t *testing.T) {
	m := jwt.NewJWTManager("access-secret", "refresh-secret", time.Minute, time.Hour)
	tokenString, err := m.GenerateRefreshToken("user-123")
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)
}

// TestJWT_ParseRefreshToken_Success проверяет корректность парсинга валидного refresh токена.
// Тестирует:
// - Успешное извлечение claims из сгенерированного токена
// - Соответствие идентификатора пользователя
// - Корректность срока действия токена (с допуском ±2 секунды)
func TestJWT_ParseRefreshToken_Success(t *testing.T) {
	m := jwt.NewJWTManager("access-secret", "refresh-secret", time.Minute, time.Hour)

	tokenString, err := m.GenerateRefreshToken("user-123")

	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	claims, err := m.ParseRefreshToken(tokenString)

	require.NoError(t, err)
	require.Equal(t, "user-123", claims.UserID)
	require.WithinDuration(t, time.Now().Add(time.Hour), claims.ExpiresAt.Time, time.Second*2)
}

// TestJWT_ParseRefreshToken_Error проверяет обработку невалидных refresh токенов.
// Сценарии:
// - Парсинг заведомо поврежденной строки токена
// - Парсинг токена, подписанного неверным секретом
// Ожидается:
// - Возврат ошибки во всех проблемных случаях
func TestJWT_ParseRefreshToken_Error(t *testing.T) {
	m := jwt.NewJWTManager("access-secret", "refresh-secret", time.Minute, time.Hour)

	_, err := m.ParseRefreshToken("invalid.token.string")

	require.Error(t, err)

	tokenString, err := m.GenerateRefreshToken("user-123")

	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	badM := jwt.NewJWTManager("access-secret", "bad-secret", time.Minute, time.Hour)

	_, err = badM.ParseRefreshToken(tokenString)
	require.Error(t, err)

}
