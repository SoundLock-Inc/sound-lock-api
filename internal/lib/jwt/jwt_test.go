package jwt_test

import (
	"sound_lock/internal/lib/jwt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJWT_GenerateAccessToken_Success(t *testing.T) {
	m := jwt.NewJWTManager("access-secret", "refresh-secret", time.Minute, time.Hour)
	tokenString, err := m.GenerateAccessToken("user-123")
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)
}

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

func TestJWT_GenerateRefreshToken_Success(t *testing.T) {
	m := jwt.NewJWTManager("access-secret", "refresh-secret", time.Minute, time.Hour)
	tokenString, err := m.GenerateRefreshToken("user-123")
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)
}

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
