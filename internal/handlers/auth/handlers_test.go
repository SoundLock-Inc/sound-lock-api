package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sound_lock/internal/domain/models"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthService — мок реализации AuthService для тестов.
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) LoginUser(ctx context.Context, email, password string) (models.Tokens, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(models.Tokens), args.Error(1)
}

func (m *MockAuthService) RegisterUser(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) LogoutUser(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) RefreshTokens(ctx context.Context, refreshToken string) (models.Tokens, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(models.Tokens), args.Error(1)
}

// TestHandlers_Login_Succes проверяет успешный сценарий логина.
// Отправляется корректный JSON, мок возвращает токены, ожидается статус 200 и корректный JSON в ответе.
func TestHandlers_Login_Succes(t *testing.T) {
	mockAuthService := new(MockAuthService)
	h := newForTest(mockAuthService)
	e := echo.New()

	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"email":"test@example.com","password":"12345"}`

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	expectedTokens := models.Tokens{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	var tokens models.Tokens
	mockAuthService.On("LoginUser", c.Request().Context(), "test@example.com", "12345").Return(expectedTokens, nil)

	h.login(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tokens))
	require.Equal(t, expectedTokens, tokens)
}

// TestHandlers_Login_ErrorLoginUser проверяет поведение при ошибке в AuthService.LoginUser.
// Ожидается возврат статуса 500.
func TestHandlers_Login_ErrorLoginUser(t *testing.T) {
	mockAuthSerice := new(MockAuthService)
	h := newForTest(mockAuthSerice)
	e := echo.New()

	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"email":"test@example.com","password":"12345"}`

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	mockAuthSerice.On("LoginUser", c.Request().Context(), "test@example.com", "12345").Return(models.Tokens{}, errors.New("Error login user"))

	h.login(c)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestHandlers_Login_ErrorParseRequest проверяет реакцию на некорректный JSON в теле запроса.
// Ожидается возврат статуса 400.
func TestHandlers_Login_ErrorParseRequest(t *testing.T) {
	mockAuthSerice := new(MockAuthService)
	h := newForTest(mockAuthSerice)
	e := echo.New()

	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"email:"test@example.com","password":"12345"}`

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	h.login(c)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestHandlers_Register_Success проверяет успешную регистрацию нового пользователя.
// Ожидается возврат статуса 201 и ID созданного пользователя в JSON.
func TestHandlers_Register_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	h := newForTest(mockAuthService)
	e := echo.New()

	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"email":"test@example.com","password":"12345"}`
	userID := uuid.New().String()

	type response struct {
		ID string `json:"id"`
	}
	expectedJson := response{
		ID: userID,
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	var r response

	mockAuthService.On("RegisterUser", c.Request().Context(), "test@example.com", "12345").Return(userID, nil)

	h.register(c)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &r))
	require.Equal(t, expectedJson, r)
}

// TestHandlers_Register_ErrorRegisterUser проверяет поведение при ошибке AuthService.RegisterUser.
// Ожидается возврат статуса 500.
func TestHandlers_Register_ErrorRegisterUser(t *testing.T) {
	mockAuthService := new(MockAuthService)
	h := newForTest(mockAuthService)
	e := echo.New()

	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"email":"test@example.com","password":"12345"}`

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	mockAuthService.On("RegisterUser", c.Request().Context(), "test@example.com", "12345").Return("", errors.New("Error register user"))

	h.register(c)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestHandlers_Register_ErrorParseRegisterRequest проверяет реакцию на некорректный JSON при регистрации.
// Ожидается возврат статуса 400.
func TestHandlers_Register_ErrorParseRegisterRequest(t *testing.T) {
	mockAuthService := new(MockAuthService)
	h := newForTest(mockAuthService)
	e := echo.New()

	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"email:"test@example.com","password":"12345"}`

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	h.register(c)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestHandlers_Logout_Success проверяет успешный выход пользователя из системы.
// Ожидается возврат статуса 200 и пустой ответ.
func TestHandlers_Logout_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	h := newForTest(mockAuthService)
	e := echo.New()
	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"refresh":"refresh-token"}`

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	mockAuthService.On("LogoutUser", c.Request().Context(), "refresh-token").Return(nil)

	h.logout(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "", rec.Body.String())
}

// TestHandlers_Logout_ErrorLogoutUser проверяет поведение при ошибке AuthService.LogoutUser.
// Ожидается возврат статуса 500.
func TestHandlers_Logout_ErrorLogoutUser(t *testing.T) {
	mockAuthService := new(MockAuthService)
	h := newForTest(mockAuthService)
	e := echo.New()
	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"refresh":"refresh-token"}`

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	mockAuthService.On("LogoutUser", c.Request().Context(), "refresh-token").Return(errors.New("Error logout user"))

	h.logout(c)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestHandlers_Logout_ErrorParseLogoutRequest проверяет реакцию на некорректный JSON при логауте.
// Ожидается возврат статуса 400.
func TestHandlers_Logout_ErrorParseLogoutRequest(t *testing.T) {
	mockAuthService := new(MockAuthService)
	h := newForTest(mockAuthService)
	e := echo.New()
	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"refresh:"refresh-token"}`

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	h.logout(c)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestHandlers_RefreshToken_Success проверяет успешное обновление токена.
// Ожидается возврат статуса 200 и новый набор токенов в JSON.
func TestHandlers_RefreshToken_Success(t *testing.T) {
	mockAuthSercie := new(MockAuthService)
	h := newForTest(mockAuthSercie)
	e := echo.New()
	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"refresh":"refresh-token"}`

	req := httptest.NewRequest(http.MethodPatch, "/auth/refresh", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	expectedTokens := models.Tokens{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}
	var tokens models.Tokens
	mockAuthSercie.On("RefreshTokens", c.Request().Context(), "refresh-token").Return(expectedTokens, nil)

	h.refresh(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tokens))
	require.Equal(t, expectedTokens, tokens)
}

// TestHandlers_RefreshToken_ErrorRefreshToken проверяет поведение при ошибке AuthService.RefreshTokens.
// Ожидается возврат статуса 500.
func TestHandlers_RefreshToken_ErrorRefreshToken(t *testing.T) {
	mockAuthSercie := new(MockAuthService)
	h := newForTest(mockAuthSercie)
	e := echo.New()
	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"refresh":"refresh-token"}`

	req := httptest.NewRequest(http.MethodPatch, "/auth/refresh", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	mockAuthSercie.On("RefreshTokens", c.Request().Context(), "refresh-token").Return(models.Tokens{}, errors.New("Error refresh token"))

	h.refresh(c)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestHandlers_RefreshToken_ErrorParseRefreshTokenRequest проверяет реакцию на некорректный JSON при обновлении токена.
// Ожидается возврат статуса 400.
func TestHandlers_RefreshToken_ErrorParseRefreshTokenRequest(t *testing.T) {
	mockAuthSercie := new(MockAuthService)
	h := newForTest(mockAuthSercie)
	e := echo.New()
	group := e.Group("/auth")
	h.SetupAuthHandlers(group)

	body := `{"refresh:"refresh-token"}`

	req := httptest.NewRequest(http.MethodPatch, "/auth/refresh", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	h.refresh(c)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
