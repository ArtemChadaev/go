package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ArtemChadaev/go/pkg/api"
	"github.com/ArtemChadaev/go/pkg/models"
	"github.com/ArtemChadaev/go/pkg/service"
	"github.com/ArtemChadaev/go/tests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// mockRedisClient - простой мок для Redis клиента
type mockRedisClient struct {
	redis.Cmdable
}

func (m *mockRedisClient) Pipeline() redis.Pipeliner {
	return &mockPipeline{}
}

type mockPipeline struct {
	redis.Pipeliner
}

func (m *mockPipeline) Incr(ctx context.Context, key string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	cmd.SetVal(1)
	return cmd
}

func (m *mockPipeline) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(ctx)
	cmd.SetVal(true)
	return cmd
}

func (m *mockPipeline) Exec(ctx context.Context) ([]redis.Cmder, error) {
	return []redis.Cmder{}, nil
}


func TestHandler_signUp_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServices := mocks.NewMockServiceAutorization(ctrl)
	
	// Создаем сервис с правильной структурой
	services := &service.Service{
		Autorization: mockServices,
	}

	// Создаем мок Redis клиента
	redisClient := &mockRedisClient{}

	handler := api.NewHandler(services, redisClient)

	// Устанавливаем gin в тестовом режиме
	gin.SetMode(gin.TestMode)

	user := models.User{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedTokens := models.ResponseTokens{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
	}

	mockServices.EXPECT().CreateUser(user).Return(1, nil)
	mockServices.EXPECT().GenerateTokens(user.Email, user.Password).Return(expectedTokens, nil)

	// Создаем JSON тело запроса
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Используем роутер для вызова приватных методов
	router := handler.InitRoutes()
	router.ServeHTTP(w, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ResponseTokens
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTokens, response)
}

func TestHandler_signUp_InvalidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServices := mocks.NewMockServiceAutorization(ctrl)
	
	services := &service.Service{
		Autorization: mockServices,
	}

	redisClient := &mockRedisClient{}
	handler := api.NewHandler(services, redisClient)

	gin.SetMode(gin.TestMode)

	// Невалидный JSON
	req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := handler.InitRoutes()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_signUp_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServices := mocks.NewMockServiceAutorization(ctrl)
	
	services := &service.Service{
		Autorization: mockServices,
	}

	redisClient := &mockRedisClient{}
	handler := api.NewHandler(services, redisClient)

	gin.SetMode(gin.TestMode)

	user := models.User{
		Email:    "existing@example.com",
		Password: "password123",
	}

	mockServices.EXPECT().CreateUser(user).Return(0, models.ErrUserAlreadyExists)

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := handler.InitRoutes()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandler_signIn_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServices := mocks.NewMockServiceAutorization(ctrl)
	
	services := &service.Service{
		Autorization: mockServices,
	}

	redisClient := &mockRedisClient{}
	handler := api.NewHandler(services, redisClient)

	gin.SetMode(gin.TestMode)

	user := models.User{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedTokens := models.ResponseTokens{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
	}

	mockServices.EXPECT().GenerateTokens(user.Email, user.Password).Return(expectedTokens, nil)

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := handler.InitRoutes()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ResponseTokens
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTokens, response)
}

func TestHandler_signIn_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServices := mocks.NewMockServiceAutorization(ctrl)
	
	services := &service.Service{
		Autorization: mockServices,
	}

	redisClient := &mockRedisClient{}
	handler := api.NewHandler(services, redisClient)

	gin.SetMode(gin.TestMode)

	user := models.User{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockServices.EXPECT().GenerateTokens(user.Email, user.Password).Return(models.ResponseTokens{}, models.ErrInvalidCredentials)

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := handler.InitRoutes()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_updateToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServices := mocks.NewMockServiceAutorization(ctrl)
	
	services := &service.Service{
		Autorization: mockServices,
	}

	redisClient := &mockRedisClient{}
	handler := api.NewHandler(services, redisClient)

	gin.SetMode(gin.TestMode)

	tokenRequest := models.ResponseTokens{
		RefreshToken: "valid_refresh_token",
	}

	expectedTokens := models.ResponseTokens{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
	}

	mockServices.EXPECT().GetAccessToken(tokenRequest.RefreshToken).Return(expectedTokens, nil)

	body, _ := json.Marshal(tokenRequest)
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := handler.InitRoutes()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ResponseTokens
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTokens, response)
}

func TestHandler_updateToken_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockServices := mocks.NewMockServiceAutorization(ctrl)
	
	services := &service.Service{
		Autorization: mockServices,
	}

	redisClient := &mockRedisClient{}
	handler := api.NewHandler(services, redisClient)

	gin.SetMode(gin.TestMode)

	tokenRequest := models.ResponseTokens{
		RefreshToken: "invalid_refresh_token",
	}

	mockServices.EXPECT().GetAccessToken(tokenRequest.RefreshToken).Return(models.ResponseTokens{}, models.ErrInvalidToken)

	body, _ := json.Marshal(tokenRequest)
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := handler.InitRoutes()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
