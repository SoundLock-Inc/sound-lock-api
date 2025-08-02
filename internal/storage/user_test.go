package storage_test

import (
	"context"
	"errors"
	"sound_lock/internal/domain/models"
	"sound_lock/internal/storage"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestStorage_CreateUser_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userStorage := storage.NewUser(mock)

	email := "test@example.com"
	password := "hashedPassword"
	userID := uuid.New()

	mock.ExpectQuery("insert into user").WithArgs(email, password).WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(userID))

	id, err := userStorage.CreateUser(context.Background(), email, password)
	require.NoError(t, err)
	require.Equal(t, userID.String(), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStorage_CreateUser_ErrorInsert(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userStorage := storage.NewUser(mock)

	email := "test@example.com"
	password := "hashedPassword"

	mock.ExpectQuery("insert into user").WithArgs(email, password).WillReturnError(errors.New("error insert user"))

	id, err := userStorage.CreateUser(context.Background(), email, password)
	require.Error(t, err)
	require.Empty(t, id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStorage_ReadUser_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userStorage := storage.NewUser(mock)

	email := "test@example.com"
	password := "hashedPassword"
	userID := uuid.New()

	mock.ExpectQuery("select id, email, password from users").WithArgs(email).WillReturnRows(pgxmock.NewRows([]string{"id", "email", "password"}).AddRow(userID, email, password))

	user, err := userStorage.ReadUser(context.Background(), email)

	require.NoError(t, err)
	require.Equal(t, email, user.Email)
	require.Equal(t, password, user.Password)
	require.Equal(t, userID, user.UUID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStorage_ReadUser_ErrorReadUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	userStorage := storage.NewUser(mock)

	email := "test@example.com"
	eUser := models.UserDTO{}
	mock.ExpectQuery("select id, email, password from users").WithArgs(email).WillReturnError(errors.New("error read user"))

	user, err := userStorage.ReadUser(context.Background(), email)

	require.Error(t, err)
	require.Equal(t, eUser, user)
	require.NoError(t, mock.ExpectationsWereMet())
}
