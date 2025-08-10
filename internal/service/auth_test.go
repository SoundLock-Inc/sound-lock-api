package service

import (
	"context"
	"errors"
	"sound_lock/internal/domain/models"
	"sound_lock/internal/lib/jwt"
	"sound_lock/internal/service/mocks"
	"testing"
	"time"

	j "github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type authMocks struct {
	userSaver           *mocks.MockUserSaver
	userProvider        *mocks.MockUserProvider
	refreshTokenSaver   *mocks.MockRefreshTokenSaver
	refreshTokenUpdater *mocks.MockRefreshTokenProvider
	refreshTokenRemover *mocks.MockRefreshTokenRemover
	jwtMethods          *mocks.MockJWTMethods
}

func setupAuthMocks(ctrl *gomock.Controller) (*Auth, *authMocks) {
	mocks := &authMocks{
		userSaver:           mocks.NewMockUserSaver(ctrl),
		userProvider:        mocks.NewMockUserProvider(ctrl),
		refreshTokenSaver:   mocks.NewMockRefreshTokenSaver(ctrl),
		refreshTokenUpdater: mocks.NewMockRefreshTokenProvider(ctrl),
		refreshTokenRemover: mocks.NewMockRefreshTokenRemover(ctrl),
		jwtMethods:          mocks.NewMockJWTMethods(ctrl),
	}

	auth := newAuthTest(mocks.userProvider, mocks.userSaver, mocks.refreshTokenSaver, mocks.refreshTokenUpdater, mocks.refreshTokenRemover, mocks.jwtMethods, 15*time.Minute, 7*24*time.Hour)

	return auth, mocks
}

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

	auth, mocks := setupAuthMocks(ctrl)

	expectedUUID := "uuid-123"

	mocks.userSaver.EXPECT().CreateUser(gomock.Any(), email, gomock.Any()).Return(UUID, nil)

	uuid, err := auth.RegisterUser(context.Background(), email, password)

	assert.NoError(t, err)
	assert.Equal(t, expectedUUID, uuid)
}

// TestAuth_RegisterUser_CreateUserError проверяет поведение при ошибке создания пользователя.
// Ожидается ошибка и пустой UUID.
func TestAuth_RegisterUser_CreateUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	auth, mocks := setupAuthMocks(ctrl)

	mocks.userSaver.EXPECT().CreateUser(gomock.Any(), email, gomock.Any()).Return("", errors.New("Error create user"))

	uuid, err := auth.RegisterUser(context.Background(), email, password)

	assert.Error(t, err)
	assert.Empty(t, uuid)
}

// TestAuth_LoginUser_Success проверяет успешный вход пользователя с корректным паролем.
// Ожидается генерация access и refresh токенов и их успешное сохранение.
func TestAuth_LoginUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	auth, mocks := setupAuthMocks(ctrl)

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := models.UserDTO{UUID: userID, Email: email, Password: string(hashedPassword)}

	accessToken := "access-token"
	refreshToken := "refresh-token"

	mocks.userProvider.EXPECT().ReadUser(gomock.Any(), email).Return(testUser, nil)
	mocks.jwtMethods.EXPECT().GenerateAccessToken(userID.String()).Return(accessToken, nil)
	mocks.jwtMethods.EXPECT().GenerateRefreshToken(userID.String()).Return(refreshToken, nil)
	mocks.refreshTokenSaver.EXPECT().CreateRefreshToken(gomock.Any(), userID.String(), refreshToken, gomock.Any()).Return(nil)

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

	auth, mocks := setupAuthMocks(ctrl)

	mocks.userProvider.EXPECT().ReadUser(gomock.Any(), email).Return(models.UserDTO{}, errors.New("Error read user"))

	tokens, err := auth.LoginUser(context.Background(), email, password)

	assert.Error(t, err)
	assert.Equal(t, models.Tokens{}, tokens)
}

// TestAuth_LoginUser_ErrorGenerateAccessToken проверяет ситуацию,
// когда не удалось сгенерировать access токен. Ожидается ошибка и пустой access токен.
func TestAuth_LoginUser_ErrorGenerateAccessToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	auth, mocks := setupAuthMocks(ctrl)

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := models.UserDTO{UUID: userID, Email: email, Password: string(hashedPassword)}

	mocks.userProvider.EXPECT().ReadUser(gomock.Any(), email).Return(testUser, nil)
	mocks.jwtMethods.EXPECT().GenerateAccessToken(userID.String()).Return("", errors.New("Error access token generate"))

	tokens, err := auth.LoginUser(context.Background(), email, password)

	assert.Error(t, err)
	assert.Empty(t, tokens.AccessToken)
}

// TestAuth_LoginUser_ErrorGenerateRefreshToken проверяет ситуацию,
// когда не удалось сгенерировать refresh токен. Ожидается ошибка и пустой refresh токен.
func TestAuth_LoginUser_ErrorGenerateRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	auth, mocks := setupAuthMocks(ctrl)

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := models.UserDTO{UUID: userID, Email: email, Password: string(hashedPassword)}
	accessToken := "access-token"

	mocks.userProvider.EXPECT().ReadUser(gomock.Any(), email).Return(testUser, nil)

	mocks.jwtMethods.EXPECT().GenerateAccessToken(userID.String()).Return(accessToken, nil)
	mocks.jwtMethods.EXPECT().GenerateRefreshToken(userID.String()).Return("", errors.New("Error generate refresh token"))

	tokens, err := auth.LoginUser(context.Background(), email, password)

	assert.Error(t, err)
	assert.Empty(t, tokens.RefreshToken)
}

// TestAuth_LoginUser_ErrorCreateRefreshToken проверяет ситуацию,
// когда не удалось сохранить refresh токен в хранилище. Ожидается ошибка и пустые токены.
func TestAuth_LoginUser_ErrorCreateRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	auth, mocks := setupAuthMocks(ctrl)

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := models.UserDTO{UUID: userID, Email: email, Password: string(hashedPassword)}
	accessToken := "access-token"
	refreshToken := "refresh-token"

	mocks.jwtMethods.EXPECT().GenerateAccessToken(userID.String()).Return(accessToken, nil)
	mocks.jwtMethods.EXPECT().GenerateRefreshToken(userID.String()).Return(refreshToken, nil)
	mocks.userProvider.EXPECT().ReadUser(gomock.Any(), email).Return(testUser, nil)
	mocks.refreshTokenSaver.EXPECT().CreateRefreshToken(gomock.Any(), userID.String(), gomock.Any(), gomock.Any()).Return(errors.New("Error create refreshToken"))

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

	auth, mocks := setupAuthMocks(ctrl)

	token := "refresh-token"

	mocks.jwtMethods.EXPECT().ParseRefreshToken(token).Return(&jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: UUID}, nil)
	mocks.refreshTokenRemover.EXPECT().DeleteRefreshToken(gomock.Any(), UUID, token).Return(nil)

	err := auth.LogoutUser(context.Background(), token)

	assert.NoError(t, err)
}

// TestAuth_Logout_ErrorParseToken проверяет ситуацию, когда не удалось распарсить refresh токен.
// Ожидается ошибка выхода.
func TestAuth_Logout_ErrorParseToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	auth, mocks := setupAuthMocks(ctrl)

	token := "refresh-token"

	mocks.jwtMethods.EXPECT().ParseRefreshToken(token).Return(&jwt.TokenClaims{}, errors.New("Error parse token"))

	err := auth.LogoutUser(context.Background(), token)

	assert.Error(t, err)
}

// TestAuth_Logout_ErrorRemoveToken проверяет ситуацию, когда удаление refresh токена завершилось ошибкой.
// Ожидается ошибка выхода.
func TestAuth_Logout_ErrorRemoveToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	auth, mocks := setupAuthMocks(ctrl)

	token := "refresh-token"
	userID := uuid.New().String()

	mocks.jwtMethods.EXPECT().ParseRefreshToken(token).Return(&jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: userID}, nil)
	mocks.refreshTokenRemover.EXPECT().DeleteRefreshToken(gomock.Any(), userID, token).Return(errors.New("Error remove token"))

	err := auth.LogoutUser(context.Background(), token)

	assert.Error(t, err)
}

// TestAuth_RefreshToken_Success проверяет успешное обновление токенов.
// Ожидается успешная генерация новых токенов и обновление refresh токена.
func TestAuth_RefreshToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	auth, mocks := setupAuthMocks(ctrl)

	refreshToken := "refresh-token"
	accessToken := "access-token"
	userID := uuid.New().String()
	tokenClaim := &jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: userID}

	mocks.jwtMethods.EXPECT().ParseRefreshToken(refreshToken).Return(tokenClaim, nil)
	mocks.jwtMethods.EXPECT().GenerateAccessToken(userID).Return(accessToken, nil)
	mocks.jwtMethods.EXPECT().GenerateRefreshToken(userID).Return(refreshToken, nil)
	mocks.refreshTokenUpdater.EXPECT().UpdateRefreshToken(gomock.Any(), userID, refreshToken, refreshToken, gomock.Any()).Return(nil)

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

	auth, mocks := setupAuthMocks(ctrl)

	refreshToken := "refresh-token"
	accessToken := "access-token"
	userID := uuid.New().String()
	tokenClaim := &jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: userID}

	mocks.jwtMethods.EXPECT().ParseRefreshToken(refreshToken).Return(tokenClaim, nil)
	mocks.jwtMethods.EXPECT().GenerateAccessToken(userID).Return(accessToken, nil)
	mocks.jwtMethods.EXPECT().GenerateRefreshToken(userID).Return(refreshToken, nil)
	mocks.refreshTokenUpdater.EXPECT().UpdateRefreshToken(gomock.Any(), userID, refreshToken, refreshToken, gomock.Any()).Return(errors.New("Error update token"))

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

	auth, mocks := setupAuthMocks(ctrl)

	refreshToken := "refresh-token"
	accessToken := "access-token"
	userID := uuid.New().String()
	tokenClaim := &jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: userID}

	mocks.jwtMethods.EXPECT().ParseRefreshToken(refreshToken).Return(tokenClaim, nil)
	mocks.jwtMethods.EXPECT().GenerateAccessToken(userID).Return(accessToken, nil)
	mocks.jwtMethods.EXPECT().GenerateRefreshToken(userID).Return("", errors.New("Error generate refresh token"))

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

	auth, mocks := setupAuthMocks(ctrl)

	refreshToken := "refresh-token"
	userID := uuid.New().String()
	tokenClaim := &jwt.TokenClaims{RegisteredClaims: j.RegisteredClaims{}, UserID: userID}

	mocks.jwtMethods.EXPECT().ParseRefreshToken(refreshToken).Return(tokenClaim, nil)
	mocks.jwtMethods.EXPECT().GenerateAccessToken(userID).Return("", errors.New("Error generate access token"))

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

	auth, mocks := setupAuthMocks(ctrl)

	refreshToken := "refresh-token"
	tokenClaim := &jwt.TokenClaims{}

	mocks.jwtMethods.EXPECT().ParseRefreshToken(refreshToken).Return(tokenClaim, errors.New("Error parse token"))

	tokens, err := auth.RefreshTokens(context.Background(), refreshToken)

	assert.Error(t, err)
	assert.Empty(t, tokens.AccessToken)
	assert.Empty(t, tokens.RefreshToken)
}
