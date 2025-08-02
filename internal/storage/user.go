package storage

import (
	"context"
	"fmt"
	"sound_lock/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserExecutor interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// User - структура для работы с таблицами пользователей и авторизацией.
// Хранит подключение к базе данных (pgxpool).
type User struct {
	db UserExecutor
}

// NewUser - конструктор, возвращает новый экземпляр Auth с подключением к базе данных.
func NewUser(db UserExecutor) *User {
	return &User{
		db: db,
	}
}

// CreateUser - создает нового пользователя в таблице users.
// На вход принимает email и захешированный пароль.
// Возвращает UUID созданного пользователя в виде строки.
// Если при выполнении запроса произошла ошибка, возвращает её наружу.
func (a *User) CreateUser(ctx context.Context, email, password string) (string, error) {
	var id uuid.UUID

	query := "insert into users (email, password) values ($1, $2) returning id"

	err := a.db.QueryRow(ctx, query, email, password).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to insert user: %w", err)
	}

	return id.String(), nil
}

// ReadUser - возвращает данные пользователя по его email.
// Достает id, email и пароль (захешированный) из базы данных.
// Возвращает структуру UserDTO или ошибку, если пользователь не найден.
func (a *User) ReadUser(ctx context.Context, email string) (models.UserDTO, error) {
	var user models.UserDTO

	query := "select id, email, password from users where email=$1"

	err := a.db.QueryRow(ctx, query, email).Scan(&user.UUID, &user.Email, &user.Password)
	if err != nil {
		return models.UserDTO{}, fmt.Errorf("failed: %w", err)
	}

	return user, nil
}
