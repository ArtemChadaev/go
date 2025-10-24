package unit

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/ArtemChadaev/go/pkg/models"
	"github.com/ArtemChadaev/go/pkg/service"
	"github.com/ArtemChadaev/go/tests/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

// Копия структуры tokenClaims из auth.go для тестов
type tokenClaims struct {
	jwt.RegisteredClaims
	UserId int `json:"user_id"`
}

func TestAuthService_CreateUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	user := models.User{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedUserID := 1

	// Ожидаем вызов CreateUser с захешированным паролем
	mockRepo.EXPECT().CreateUser(gomock.Any()).Return(expectedUserID, nil)

	// Ожидаем вызов CreateInitialUserSettings
	mockSettingsService.EXPECT().CreateInitialUserSettings(expectedUserID, "test").Return(nil)

	userID, err := authService.CreateUser(user)

	assert.NoError(t, err)
	assert.Equal(t, expectedUserID, userID)
}

func TestAuthService_CreateUser_EmailAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	user := models.User{
		Email:    "existing@example.com",
		Password: "password123",
	}

	pqErr := &pq.Error{Code: "23505"}
	mockRepo.EXPECT().CreateUser(gomock.Any()).Return(0, pqErr)

	userID, err := authService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, models.ErrUserAlreadyExists, err)
	assert.Equal(t, 0, userID)
}

func TestAuthService_CreateUser_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	user := models.User{
		Email:    "test@example.com",
		Password: "password123",
	}

	internalErr := errors.New("database error")
	mockRepo.EXPECT().CreateUser(gomock.Any()).Return(0, internalErr)

	userID, err := authService.CreateUser(user)

	assert.Error(t, err)
	assert.IsType(t, &models.AppError{}, err)
	assert.Equal(t, 0, userID)
}

func TestAuthService_GenerateTokens_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	email := "test@example.com"
	password := "password123"
	userID := 1

	mockRepo.EXPECT().GetUser(email, gomock.Any()).Return(userID, nil)
	mockRepo.EXPECT().CreateToken(gomock.Any()).Return(nil)

	tokens, err := authService.GenerateTokens(email, password)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
}

func TestAuthService_GenerateTokens_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	email := "test@example.com"
	password := "wrongpassword"

	mockRepo.EXPECT().GetUser(email, gomock.Any()).Return(0, sql.ErrNoRows)

	tokens, err := authService.GenerateTokens(email, password)

	assert.Error(t, err)
	assert.Equal(t, models.ErrInvalidCredentials, err)
	assert.Empty(t, tokens.AccessToken)
	assert.Empty(t, tokens.RefreshToken)
}

func TestAuthService_GetAccessToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	refreshToken := "valid_refresh_token"
	userID := 1

	refreshTokenData := models.RefreshToken{
		ID:        1,
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}

	mockRepo.EXPECT().GetRefreshToken(refreshToken).Return(refreshTokenData, nil)
	mockRepo.EXPECT().UpdateToken(refreshToken, gomock.Any()).Return(nil).AnyTimes()

	tokens, err := authService.GetAccessToken(refreshToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
}

func TestAuthService_GetAccessToken_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	refreshToken := "invalid_token"

	mockRepo.EXPECT().GetRefreshToken(refreshToken).Return(models.RefreshToken{}, sql.ErrNoRows)

	tokens, err := authService.GetAccessToken(refreshToken)

	assert.Error(t, err)
	assert.Equal(t, models.ErrInvalidToken, err)
	assert.Empty(t, tokens.AccessToken)
	assert.Empty(t, tokens.RefreshToken)
}

func TestAuthService_GetAccessToken_ExpiredToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	refreshToken := "expired_token"
	userID := 1

	refreshTokenData := models.RefreshToken{
		ID:        1,
		UserID:    userID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(-time.Hour), // Истекший токен
	}

	mockRepo.EXPECT().GetRefreshToken(refreshToken).Return(refreshTokenData, nil)
	mockRepo.EXPECT().DeleteRefreshToken(refreshTokenData.ID).Return(nil)

	tokens, err := authService.GetAccessToken(refreshToken)

	assert.Error(t, err)
	assert.Equal(t, models.ErrInvalidToken, err)
	assert.Empty(t, tokens.AccessToken)
	assert.Empty(t, tokens.RefreshToken)
}

func TestAuthService_ParseToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	// Создаем валидный токен для теста
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserId: 123,
	})

	accessToken, err := token.SignedString([]byte("awsg8s#@4Sf86DS#$2dF"))
	assert.NoError(t, err)

	userID, err := authService.ParseToken(accessToken)

	assert.NoError(t, err)
	assert.Equal(t, 123, userID)
}

func TestAuthService_ParseToken_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	invalidToken := "invalid.jwt.token"

	userID, err := authService.ParseToken(invalidToken)

	assert.Error(t, err)
	assert.Equal(t, models.ErrInvalidToken, err)
	assert.Equal(t, 0, userID)
}

func TestAuthService_UnAuthorize_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	refreshToken := "valid_token"
	refreshTokenData := models.RefreshToken{
		ID:     1,
		UserID: 1,
		Token:  refreshToken,
	}

	mockRepo.EXPECT().GetRefreshToken(refreshToken).Return(refreshTokenData, nil)
	mockRepo.EXPECT().DeleteRefreshToken(refreshTokenData.ID).Return(nil)

	err := authService.UnAuthorize(refreshToken)

	assert.NoError(t, err)
}

func TestAuthService_UnAuthorizeAll_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageAutorization(ctrl)
	mockSettingsService := mocks.NewMockServiceUserSettings(ctrl)

	authService := service.NewAuthService(mockRepo, mockSettingsService)

	email := "test@example.com"
	password := "password123"
	userID := 1

	mockRepo.EXPECT().GetUser(email, password).Return(userID, nil)
	mockRepo.EXPECT().DeleteAllUserRefreshTokens(userID).Return(nil)

	err := authService.UnAuthorizeAll(email, password)

	assert.NoError(t, err)
}
