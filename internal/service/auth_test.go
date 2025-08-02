package service_test

import (
	"context"
	"errors"
	"sound_lock/internal/domain/models"
	"sound_lock/internal/lib/jwt"
	"sound_lock/internal/service"
	"sound_lock/internal/service/mocks"
	"testing"
	"time"

	j "github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

const (
	email    = "test@example.com"
	password = "pass123"
	UUID     = "uuid-123"
)

// TestAuth_RegisterUser_Success проверяет успешную регистрацию пользователя.
// Ожидается, что CreateUser вернет корректный UUID и ошибок не будет.
func TestAuth_RegisterUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedUUID := "uuid-123"
	mockUserSaver := mocks.NewMockUserSaver(ctrl)

	mockUserSaver.EXPECT().CreateUser(gomock.Any(), email, gomock.Any()).Return(UUID, nil)

	auth := &service.Auth{
		UserSaver: mockUserSaver,
	}

	uuid, err := auth.RegisterUser(context.Background(), email, password)

	assert.NoError(t, err)
	assert.Equal(t, expectedUUID, uuid)
}

// TestAuth_RegisterUser_CreateUserError проверяет поведение при ошибке создания пользователя.
// Ожидается ошибка и пустой UUID.
func TestAuth_RegisterUser_CreateUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserSaver := mocks.NewMockUserSaver(ctrl)

	mockUserSaver.EXPECT().CreateUser(gomock.Any(), email, gomock.Any()).Return("", errors.New("Error create user"))

	auth := &service.Auth{
		UserSaver: mockUserSaver,
	}

	uuid, err := auth.RegisterUser(context.Background(), email, password)

	assert.Error(t, err)
	assert.Empty(t, uuid)
}

// TestAuth_LoginUser_Success проверяет успешный вход пользователя с корректным паролем.
// Ожидается генерация access и refresh токенов и их успешное сохранение.
func TestAuth_LoginUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := models.UserDTO{UUID: userID, Email: email, Password: string(hashedPassword)}

	accessToken := "access-token"
	refreshToken := "refresh-token"

	mockUserProvider := mocks.NewMockUserProvider(ctrl)
	mockRefreshTokenSaver := mocks.NewMockRefreshTokenSaver(ctrl)
	mockJwt := mocks.NewMockJWTMethods(ctrl)

	mockUserProvider.EXPECT().ReadUser(gomock.Any(), email).Return(testUser, nil)
	mockJwt.EXPECT().GenerateAccessToken(userID.String()).Return(accessToken, nil)
	mockJwt.EXPECT().GenerateRefreshToken(userID.String()).Return(refreshToken, nil)
	mockRefreshTokenSaver.EXPECT().CreateRefreshToken(gomock.Any(), userID.String(), refreshToken, gomock.Any()).Return(nil)

	auth := &service.Auth{
		UserProvider:      mockUserProvider,
		RefreshTokenSaver: mockRefreshTokenSaver,
		Jwt:               mockJwt,
	}

	tokens, err := auth.LoginUser(context.Background(), email, password)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
}

// TestAuth_LoginUser_ErrorReadUser проверяет ситуацию, когда не удалось прочитать данные пользователя из БД.
// Ожидается ошибка и пустой результат токенов.
func TestAuth_LoginUser_ErrorReadUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserProvider := mocks.NewMockUserProvider(ctrl)

	mockUserProvider.EXPECT().ReadUser(gomock.Any(), email).Return(models.UserDTO{}, errors.New("Error read user"))
	jwtManager := jwt.NewJWTManager("access-secret", "refresh-secret", time.Minute, time.Hour)

	auth := &service.Auth{UserProvider: mockUserProvider, Jwt: jwtManager}

	tokens, err := auth.LoginUser(context.Background(), email, password)

	assert.Error(t, err)
	assert.Equal(t, models.Tokens{}, tokens)
}

// TestAuth_LoginUser_ErrorGenerateAccessToken проверяет ситуацию,
// когда не удалось сгенерировать access токен. Ожидается ошибка и пустой access токен.
func TestAuth_LoginUser_ErrorGenerateAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := models.UserDTO{UUID: userID, Email: email, Password: string(hashedPassword)}

	mockUserProvider := mocks.NewMockUserProvider(ctrl)
	mockJwt := mocks.NewMockJWTMethods(ctrl)

	mockUserProvider.EXPECT().ReadUser(gomock.Any(), email).Return(testUser, nil)
	mockJwt.EXPECT().GenerateAccessToken(userID.String()).Return("", errors.New("Error access token generate"))

	auth := &service.Auth{UserProvider: mockUserProvider, Jwt: mockJwt}

	tokens, err := auth.LoginUser(context.Background(), email, password)

	assert.Error(t, err)
	assert.Empty(t, tokens.AccessToken)
}

// TestAuth_LoginUser_ErrorGenerateRefreshToken проверяет ситуацию,
// когда не удалось сгенерировать refresh токен. Ожидается ошибка и пустой refresh токен.
func TestAuth_LoginUser_ErrorGenerateRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := models.UserDTO{UUID: userID, Email: email, Password: string(hashedPassword)}
	accessToken := "access-token"

	mockUserProvider := mocks.NewMockUserProvider(ctrl)
	mockJwt := mocks.NewMockJWTMethods(ctrl)

	mockUserProvider.EXPECT().ReadUser(gomock.Any(), email).Return(testUser, nil)

	mockJwt.EXPECT().GenerateAccessToken(userID.String()).Return(accessToken, nil)
	mockJwt.EXPECT().GenerateRefreshToken(userID.String()).Return("", errors.New("Error generate refresh token"))

	auth := &service.Auth{UserProvider: mockUserProvider, Jwt: mockJwt}

	tokens, err := auth.LoginUser(context.Background(), email, password)

	assert.Error(t, err)
	assert.Empty(t, tokens.RefreshToken)
}

// TestAuth_LoginUser_ErrorCreateRefreshToken проверяет ситуацию,
// когда не удалось сохранить refresh токен в хранилище. Ожидается ошибка и пустые токены.
func TestAuth_LoginUser_ErrorCreateRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := models.UserDTO{UUID: userID, Email: email, Password: string(hashedPassword)}

	mockUserProvider := mocks.NewMockUserProvider(ctrl)
	mockRefreshTokenSaver := mocks.NewMockRefreshTokenSaver(ctrl)

	mockUserProvider.EXPECT().ReadUser(gomock.Any(), email).Return(testUser, nil)
	mockRefreshTokenSaver.EXPECT().CreateRefreshToken(gomock.Any(), userID.String(), gomock.Any(), gomock.Any()).Return(errors.New("Error create refreshToken"))

	jwtManager := jwt.NewJWTManager("access-secret", "refresh-secret", time.Minute, time.Hour)

	auth := &service.Auth{UserProvider: mockUserProvider, RefreshTokenSaver: mockRefreshTokenSaver, Jwt: jwtManager}

	tokens, err := auth.LoginUser(context.Background(), email, password)

	assert.Error(t, err)
	assert.Empty(t, tokens.AccessToken)
	assert.Empty(t, tokens.RefreshToken)
}

// TestAuth_LogoutUser_Success проверяет успешный выход пользователя.
// Ожидается успешный парсинг refresh токена и его удаление.
func TestAuth_LogoutUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	token := "refresh-token"
	mockJwt := mocks.NewMockJWTMethods(ctrl)
	mockRefreshTokenRemover := mocks.NewMockRefreshTokenRemover(ctrl)

	mockJwt.EXPECT().ParseRefreshToken(token).Return(&jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: UUID}, nil)
	mockRefreshTokenRemover.EXPECT().DeleteRefreshToken(gomock.Any(), UUID, token).Return(nil)

	auth := &service.Auth{RefreshTokenRemover: mockRefreshTokenRemover, Jwt: mockJwt}

	err := auth.LogoutUser(context.Background(), token)

	assert.NoError(t, err)
}

// TestAuth_Logout_ErrorParseToken проверяет ситуацию, когда не удалось распарсить refresh токен.
// Ожидается ошибка выхода.
func TestAuth_Logout_ErrorParseToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	token := "refresh-token"
	mockJwt := mocks.NewMockJWTMethods(ctrl)

	mockJwt.EXPECT().ParseRefreshToken(token).Return(&jwt.TokenClaims{}, errors.New("Error parse token"))

	auth := &service.Auth{Jwt: mockJwt}

	err := auth.LogoutUser(context.Background(), token)

	assert.Error(t, err)
}

// TestAuth_Logout_ErrorRemoveToken проверяет ситуацию, когда удаление refresh токена завершилось ошибкой.
// Ожидается ошибка выхода.
func TestAuth_Logout_ErrorRemoveToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	token := "refresh-token"
	userID := uuid.New().String()
	mockJwt := mocks.NewMockJWTMethods(ctrl)
	mockRefreshTokenRemover := mocks.NewMockRefreshTokenRemover(ctrl)

	mockJwt.EXPECT().ParseRefreshToken(token).Return(&jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: userID}, nil)
	mockRefreshTokenRemover.EXPECT().DeleteRefreshToken(gomock.Any(), userID, token).Return(errors.New("Error remove token"))

	auth := &service.Auth{RefreshTokenRemover: mockRefreshTokenRemover, Jwt: mockJwt}

	err := auth.LogoutUser(context.Background(), token)

	assert.Error(t, err)
}

// TestAuth_RefreshToken_Success проверяет успешное обновление токенов.
// Ожидается успешная генерация новых токенов и обновление refresh токена.
func TestAuth_RefreshToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	refreshToken := "refresh-token"
	accessToken := "access-token"
	userID := uuid.New().String()
	tokenClaim := &jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: userID}

	mockJwt := mocks.NewMockJWTMethods(ctrl)
	mockRefreshTokenProvider := mocks.NewMockRefreshTokenProvider(ctrl)

	mockJwt.EXPECT().ParseRefreshToken(refreshToken).Return(tokenClaim, nil)
	mockJwt.EXPECT().GenerateAccessToken(userID).Return(accessToken, nil)
	mockJwt.EXPECT().GenerateRefreshToken(userID).Return(refreshToken, nil)
	mockRefreshTokenProvider.EXPECT().UpdateRefreshToken(gomock.Any(), userID, refreshToken, refreshToken, gomock.Any()).Return(nil)

	auth := &service.Auth{RefreshTokenProvider: mockRefreshTokenProvider, Jwt: mockJwt}

	tokens, err := auth.RefreshTokens(context.Background(), refreshToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
}

// TestAuth_RefreshToken_ErrorUpdateToken проверяет ситуацию, когда не удалось обновить refresh токен в хранилище.
// Ожидается ошибка и пустые токены.
func TestAuth_RefreshToken_ErrorUpdateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	refreshToken := "refresh-token"
	accessToken := "access-token"
	userID := uuid.New().String()
	tokenClaim := &jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: userID}

	mockJwt := mocks.NewMockJWTMethods(ctrl)
	mockRefreshTokenProvider := mocks.NewMockRefreshTokenProvider(ctrl)

	mockJwt.EXPECT().ParseRefreshToken(refreshToken).Return(tokenClaim, nil)
	mockJwt.EXPECT().GenerateAccessToken(userID).Return(accessToken, nil)
	mockJwt.EXPECT().GenerateRefreshToken(userID).Return(refreshToken, nil)
	mockRefreshTokenProvider.EXPECT().UpdateRefreshToken(gomock.Any(), userID, refreshToken, refreshToken, gomock.Any()).Return(errors.New("Error update token"))

	auth := &service.Auth{RefreshTokenProvider: mockRefreshTokenProvider, Jwt: mockJwt}

	tokens, err := auth.RefreshTokens(context.Background(), refreshToken)

	assert.Error(t, err)
	assert.Empty(t, tokens.AccessToken)
	assert.Empty(t, tokens.RefreshToken)
}

// TestAuth_RefreshToken_ErrorGenerateRefreshToken проверяет ситуацию, когда не удалось сгенерировать новый refresh токен.
// Ожидается ошибка и пустые токены.
func TestAuth_RefreshToken_ErrorGenerateRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	refreshToken := "refresh-token"
	accessToken := "access-token"
	userID := uuid.New().String()
	tokenClaim := &jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: userID}

	mockJwt := mocks.NewMockJWTMethods(ctrl)
	mockRefreshTokenProvider := mocks.NewMockRefreshTokenProvider(ctrl)

	mockJwt.EXPECT().ParseRefreshToken(refreshToken).Return(tokenClaim, nil)
	mockJwt.EXPECT().GenerateAccessToken(userID).Return(accessToken, nil)
	mockJwt.EXPECT().GenerateRefreshToken(userID).Return("", errors.New("Error generate refresh token"))

	auth := &service.Auth{RefreshTokenProvider: mockRefreshTokenProvider, Jwt: mockJwt}

	tokens, err := auth.RefreshTokens(context.Background(), refreshToken)

	assert.Error(t, err)
	assert.Empty(t, tokens.AccessToken)
	assert.Empty(t, tokens.RefreshToken)
}

// TestAuth_RefreshToken_ErrorGenerateAccessToken проверяет ситуацию, когда не удалось сгенерировать новый access токен.
// Ожидается ошибка и пустые токены.
func TestAuth_RefreshToken_ErrorGenerateAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	refreshToken := "refresh-token"
	userID := uuid.New().String()
	tokenClaim := &jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: userID}

	mockJwt := mocks.NewMockJWTMethods(ctrl)
	mockRefreshTokenProvider := mocks.NewMockRefreshTokenProvider(ctrl)

	mockJwt.EXPECT().ParseRefreshToken(refreshToken).Return(tokenClaim, nil)
	mockJwt.EXPECT().GenerateAccessToken(userID).Return("", errors.New("Error generate access token"))

	auth := &service.Auth{RefreshTokenProvider: mockRefreshTokenProvider, Jwt: mockJwt}

	tokens, err := auth.RefreshTokens(context.Background(), refreshToken)

	assert.Error(t, err)
	assert.Empty(t, tokens.AccessToken)
	assert.Empty(t, tokens.RefreshToken)
}

// TestAuth_RefreshToken_ErrorParseRefreshToken проверяет ситуацию, когда парсинг refresh токена завершился ошибкой.
// Ожидается ошибка и пустые токены.
func TestAuth_RefreshToken_ErrorParseRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	refreshToken := "refresh-token"
	tokenClaim := &jwt.TokenClaims{}

	mockJwt := mocks.NewMockJWTMethods(ctrl)
	mockRefreshTokenProvider := mocks.NewMockRefreshTokenProvider(ctrl)

	mockJwt.EXPECT().ParseRefreshToken(refreshToken).Return(tokenClaim, errors.New("Error parse token"))

	auth := &service.Auth{RefreshTokenProvider: mockRefreshTokenProvider, Jwt: mockJwt}

	tokens, err := auth.RefreshTokens(context.Background(), refreshToken)

	assert.Error(t, err)
	assert.Empty(t, tokens.AccessToken)
	assert.Empty(t, tokens.RefreshToken)
}
