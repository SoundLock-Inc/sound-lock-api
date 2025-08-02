package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type RefreshExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// RefreshTokenStorage отвечает за операции с таблицей refresh_tokens.
// Хранит подключение к базе данных (pgxpool).
type RefreshTokenStorage struct {
	db RefreshExecutor
}

// NewRefreshToken создает новый экземпляр RefreshTokenStorage с подключением к БД.
func NewRefreshToken(db RefreshExecutor) *RefreshTokenStorage {
	return &RefreshTokenStorage{
		db: db,
	}
}

// CreateRefreshToken добавляет новый refresh токен в таблицу refresh_tokens.
// Принимает ID пользователя (userID), сам токен (token) и время истечения (expiresAt).
// Возвращает ошибку, если вставка не удалась или не затронула ни одной строки.
func (r *RefreshTokenStorage) CreateRefreshToken(ctx context.Context, userID string, token string, expiresAt time.Time) error {
	query := "insert into refresh_tokens (user_id, token, expires_at) values ($1, $2, $3)"

	data, err := r.db.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token: %w", err)
	}

	if data.RowsAffected() == 0 {
		return errors.New("updated 0 rows")
	}

	return nil
}

// UpdateRefreshToken обновляет существующий refresh токен пользователя в таблице refresh_tokens.
// Обновляет токен и дату истечения по userID и старому значению токена.
// Возвращает ошибку, если обновление не удалоcь или не затронуло ни одной строки.
func (r *RefreshTokenStorage) UpdateRefreshToken(ctx context.Context, userID string, token, oldToken string, expiresAt time.Time) error {
	query := "update refresh_tokens set token=$1, expires_at=$2 where user_id=$3 and token=$4"

	data, err := r.db.Exec(ctx, query, token, expiresAt, userID, oldToken)
	if err != nil {
		return fmt.Errorf("failed to update refresh token: %w", err)
	}

	if data.RowsAffected() == 0 {
		return errors.New("updated 0 rows")
	}

	return nil
}

// DeleteRefreshToken удаляет существующий refresh токен пользователя в таблице refresh_tokens.
// Удаляет токен по userID и значению токена.
// Возвращает ошибку, если удаление не удалоcь или не затронуло ни одной строки.
func (r *RefreshTokenStorage) DeleteRefreshToken(ctx context.Context, userID, token string) error {
	query := "delete from refresh_tokens where user_id=$1 and token=$2"

	data, err := r.db.Exec(ctx, query, userID, token)
	if err != nil {
		return fmt.Errorf("failed to update refresh token: %w", err)
	}

	if data.RowsAffected() == 0 {
		return errors.New("updated 0 rows")
	}

	return nil
}
