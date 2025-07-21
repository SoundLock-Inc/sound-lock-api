package postgres

import (
	"context"
	"fmt"
	"sound_lock/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// New создает и возвращает новое подключение к пулу PostgreSQL.
// Принимает контекст выполнения и указатель на конфигурацию приложения.
func New(context context.Context, config *config.Config) *pgxpool.Pool {
	dbHost := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		config.DB.Username, // Имя пользователя
		config.DB.Password, // Пароль
		config.DB.DBHost,   // Хост базы данных
		config.DB.DBPort,   // Порт базы данных
		config.DB.DBName,   // Название базы данных
	)

	conn, err := pgxpool.New(context, dbHost)
	if err != nil {
		panic(err)
	}

	return conn
}
