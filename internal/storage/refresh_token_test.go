package storage_test

import (
	"context"
	"errors"
	"sound_lock/internal/storage"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

// TestRefreshToken_CreateRefreshToken_Success проверяет успешное создание refresh-токена.
// Используется pgxmock для имитации успешного INSERT-запроса в таблицу refresh_tokens.
// Ожидается, что метод CreateRefreshToken не вернет ошибок и все ожидания pgxmock будут выполнены.
func TestRefreshToken_CreateRefreshToken_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()
	token := "refresh-token"
	expiresAt := time.Now().Add(time.Hour)

	refreshStorage := storage.NewRefreshToken(mock)

	mock.ExpectExec("insert into refresh_tokens").WithArgs(userID.String(), token, expiresAt).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = refreshStorage.CreateRefreshToken(context.Background(), userID.String(), token, expiresAt)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshToken_CreateRefreshToken_ErrorCreateRefreshToken проверяет обработку ошибки при выполнении INSERT-запроса.
// Мокаем ошибку выполнения запроса и ожидаем, что метод CreateRefreshToken вернет ошибку.
func TestRefreshToken_CreateRefreshToken_ErrorCreateRefreshToken(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()
	token := "refresh-token"
	expiresAt := time.Now().Add(time.Hour)

	refreshStorage := storage.NewRefreshToken(mock)

	mock.ExpectExec("insert into refresh_tokens").WithArgs(userID.String(), token, expiresAt).WillReturnError(errors.New("Error insert refresh token"))

	err = refreshStorage.CreateRefreshToken(context.Background(), userID.String(), token, expiresAt)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshToken_CreateRefreshToken_NoRowsAffected проверяет поведение при успешном выполнении INSERT,
// но без вставленных строк (затронуто 0 строк). Ожидается, что метод вернет ошибку "updated 0 rows".
func TestRefreshToken_CreateRefreshToken_NoRowsAffected(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	rtStorage := storage.NewRefreshToken(mock)

	userID := uuid.New()
	token := "refresh-token"
	expiresAt := time.Now().Add(time.Hour)

	mock.ExpectExec("insert into refresh_tokens").
		WithArgs(userID.String(), token, expiresAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 0))

	err = rtStorage.CreateRefreshToken(context.Background(), userID.String(), token, expiresAt)
	require.Error(t, err)
	require.Equal(t, "updated 0 rows", err.Error())
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshToken_UpdateRefreshToken_Success проверяет успешное обновление refresh-токена.
// Используется pgxmock для имитации успешного UPDATE-запроса. Метод не должен возвращать ошибок.
func TestRefreshToken_UpdateRefreshToken_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()
	newToken := "new-refresh-token"
	oldToken := "old-refresh-token"
	expiresAt := time.Now().Add(time.Hour)

	refreshStorage := storage.NewRefreshToken(mock)

	mock.ExpectExec("update refresh_tokens").WithArgs(newToken, expiresAt, userID.String(), oldToken).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = refreshStorage.UpdateRefreshToken(context.Background(), userID.String(), newToken, oldToken, expiresAt)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshToken_UpdateRefreshToken_ErrorUpdateRefreshToken проверяет обработку ошибки при выполнении UPDATE-запроса.
// Мокаем ошибку обновления и проверяем, что метод возвращает ошибку.
func TestRefreshToken_UpdateRefreshToken_ErrorUpdateRefreshToken(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()
	newToken := "new-refresh-token"
	oldToken := "old-refresh-token"
	expiresAt := time.Now().Add(time.Hour)

	refreshStorage := storage.NewRefreshToken(mock)

	mock.ExpectExec("update refresh_tokens").WithArgs(newToken, expiresAt, userID.String(), oldToken).WillReturnError(errors.New("Error update refresh token"))

	err = refreshStorage.UpdateRefreshToken(context.Background(), userID.String(), newToken, oldToken, expiresAt)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshToken_UpdateRefreshToken_NoRowsAffected проверяет поведение при UPDATE-запросе,
// который не обновил ни одной строки (затронуто 0 строк). Метод должен вернуть ошибку "updated 0 rows".
func TestRefreshToken_UpdateRefreshToken_NoRowsAffected(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()
	newToken := "new-refresh-token"
	oldToken := "old-refresh-token"
	expiresAt := time.Now().Add(time.Hour)

	refreshStorage := storage.NewRefreshToken(mock)

	mock.ExpectExec("update refresh_tokens").WithArgs(newToken, expiresAt, userID.String(), oldToken).WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err = refreshStorage.UpdateRefreshToken(context.Background(), userID.String(), newToken, oldToken, expiresAt)

	require.Error(t, err)
	require.Equal(t, "updated 0 rows", err.Error())
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshToken_RemoveRefreshToken_Success проверяет успешное удаление refresh-токена.
// Используется pgxmock для имитации успешного DELETE-запроса. Ошибок не должно быть.
func TestRefreshToken_RemoveRefreshToken_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()
	token := "refresh-token"

	refreshStorage := storage.NewRefreshToken(mock)

	mock.ExpectExec("delete from refresh_tokens").WithArgs(userID.String(), token).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = refreshStorage.DeleteRefreshToken(context.Background(), userID.String(), token)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshToken_RemoveRefreshToken_ErrorUpdateRefreshToken проверяет обработку ошибки при удалении токена.
// Мокаем ошибку выполнения DELETE-запроса и проверяем, что метод возвращает ошибку.
func TestRefreshToken_RemoveRefreshToken_ErrorUpdateRefreshToken(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()
	token := "refresh-token"

	refreshStorage := storage.NewRefreshToken(mock)

	mock.ExpectExec("delete from refresh_tokens").WithArgs(userID.String(), token).WillReturnError(errors.New("error delete refresh token"))

	err = refreshStorage.DeleteRefreshToken(context.Background(), userID.String(), token)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestRefreshToken_RemoveRefreshToken_NoRowsAffected проверяет поведение при DELETE-запросе,
// который не удалил ни одной строки. Метод должен вернуть ошибку "updated 0 rows".
func TestRefreshToken_RemoveRefreshToken_NoRowsAffected(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userID := uuid.New()
	token := "refresh-token"

	refreshStorage := storage.NewRefreshToken(mock)

	mock.ExpectExec("delete from refresh_tokens").WithArgs(userID.String(), token).WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err = refreshStorage.DeleteRefreshToken(context.Background(), userID.String(), token)

	require.Error(t, err)
	require.Equal(t, "updated 0 rows", err.Error())
	require.NoError(t, mock.ExpectationsWereMet())
}
