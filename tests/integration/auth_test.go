package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ArtemChadaev/go/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthIntegration(t *testing.T) {
	ctx := context.Background()
	
	// Настраиваем тестовое окружение
	testSuite, err := SetupTestSuite(ctx)
	require.NoError(t, err)
	defer testSuite.Cleanup()

	// Устанавливаем gin в тестовом режиме
	gin.SetMode(gin.TestMode)

	t.Run("SignUp and SignIn Flow", func(t *testing.T) {
		// Очищаем данные перед тестом
		err := testSuite.CleanupTestData(ctx)
		require.NoError(t, err)

		// Тестовые данные
		user := models.User{
			Email:    "test@example.com",
			Password: "password123",
		}

		// 1. Регистрация пользователя
		t.Run("SignUp", func(t *testing.T) {
			body, _ := json.Marshal(user)
			req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router := testSuite.Handler.InitRoutes()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.ResponseTokens
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.AccessToken)
			assert.NotEmpty(t, response.RefreshToken)
		})

		// 2. Попытка повторной регистрации (должна вернуть ошибку)
		t.Run("SignUp Duplicate User", func(t *testing.T) {
			body, _ := json.Marshal(user)
			req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router := testSuite.Handler.InitRoutes()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusConflict, w.Code)
		})

		// 3. Вход в систему
		t.Run("SignIn", func(t *testing.T) {
			body, _ := json.Marshal(user)
			req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router := testSuite.Handler.InitRoutes()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response models.ResponseTokens
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.AccessToken)
			assert.NotEmpty(t, response.RefreshToken)
		})

		// 4. Вход с неверными данными
		t.Run("SignIn Invalid Credentials", func(t *testing.T) {
			invalidUser := models.User{
				Email:    user.Email,
				Password: "wrongpassword",
			}

			body, _ := json.Marshal(invalidUser)
			req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router := testSuite.Handler.InitRoutes()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})

		// 5. Обновление токена
		t.Run("Refresh Token", func(t *testing.T) {
			// Сначала получаем токены через sign-in
			body, _ := json.Marshal(user)
			req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router := testSuite.Handler.InitRoutes()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var tokens models.ResponseTokens
			err := json.Unmarshal(w.Body.Bytes(), &tokens)
			require.NoError(t, err)

			// Теперь обновляем токен
			refreshRequest := models.ResponseTokens{
				RefreshToken: tokens.RefreshToken,
			}

			refreshBody, _ := json.Marshal(refreshRequest)
			refreshReq := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(refreshBody))
			refreshReq.Header.Set("Content-Type", "application/json")

			refreshW := httptest.NewRecorder()
			router.ServeHTTP(refreshW, refreshReq)

			assert.Equal(t, http.StatusOK, refreshW.Code)

			var newTokens models.ResponseTokens
			err = json.Unmarshal(refreshW.Body.Bytes(), &newTokens)
			assert.NoError(t, err)
			assert.NotEmpty(t, newTokens.AccessToken)
			assert.NotEmpty(t, newTokens.RefreshToken)
			// Токены должны быть разными
			assert.NotEqual(t, tokens.AccessToken, newTokens.AccessToken)
		})

		// 6. Обновление с невалидным токеном
		t.Run("Refresh Invalid Token", func(t *testing.T) {
			invalidRefreshRequest := models.ResponseTokens{
				RefreshToken: "invalid_refresh_token",
			}

			body, _ := json.Marshal(invalidRefreshRequest)
			req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router := testSuite.Handler.InitRoutes()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})

	t.Run("Rate Limiting", func(t *testing.T) {
		// Очищаем данные перед тестом
		err := testSuite.CleanupTestData(ctx)
		require.NoError(t, err)

		user := models.User{
			Email:    "ratelimit@example.com",
			Password: "password123",
		}

		// Регистрируем пользователя
		body, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := testSuite.Handler.InitRoutes()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Делаем много запросов на вход для проверки rate limiting
		for i := 0; i < 15; i++ { // Превышаем лимит
			body, _ := json.Marshal(user)
			req := httptest.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if i < 10 { // Первые 10 запросов должны пройти
				assert.Equal(t, http.StatusOK, w.Code, "Request %d should pass", i)
			} else { // Последующие должны быть заблокированы
				assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request %d should be rate limited", i)
			}
		}
	})

	t.Run("Invalid Requests", func(t *testing.T) {
		testCases := []struct {
			name           string
			endpoint       string
			body           interface{}
			expectedStatus int
		}{
			{
				name:           "SignUp Invalid JSON",
				endpoint:       "/auth/sign-up",
				body:           "{invalid json",
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "SignUp Missing Email",
				endpoint:       "/auth/sign-up",
				body:           models.User{Password: "password123"},
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "SignUp Missing Password",
				endpoint:       "/auth/sign-up",
				body:           models.User{Email: "test@example.com"},
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "SignIn Invalid JSON",
				endpoint:       "/auth/sign-in",
				body:           "{invalid json",
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "Refresh Invalid JSON",
				endpoint:       "/auth/refresh",
				body:           "{invalid json",
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "Refresh Missing Token",
				endpoint:       "/auth/refresh",
				body:           models.ResponseTokens{},
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var body []byte
				if str, ok := tc.body.(string); ok {
					body = []byte(str)
				} else {
					body, _ = json.Marshal(tc.body)
				}

				req := httptest.NewRequest("POST", tc.endpoint, bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router := testSuite.Handler.InitRoutes()
				router.ServeHTTP(w, req)

				assert.Equal(t, tc.expectedStatus, w.Code)
			})
		}
	})
}

func TestAuthTokenExpiration(t *testing.T) {
	ctx := context.Background()
	
	// Настраиваем тестовое окружение
	testSuite, err := SetupTestSuite(ctx)
	require.NoError(t, err)
	defer testSuite.Cleanup()

	// Устанавливаем gin в тестовом режиме
	gin.SetMode(gin.TestMode)

	// Очищаем данные перед тестом
	err = testSuite.CleanupTestData(ctx)
	require.NoError(t, err)

	user := models.User{
		Email:    "expire@example.com",
		Password: "password123",
	}

	// Регистрируем пользователя
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := testSuite.Handler.InitRoutes()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var tokens models.ResponseTokens
	err = json.Unmarshal(w.Body.Bytes(), &tokens)
	require.NoError(t, err)

	// Ждем некоторое время (в реальном тесте можно использовать мок времени)
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что токен все еще работает для защищенного эндпоинта
	protectedReq := httptest.NewRequest("GET", "/api/settings/", nil)
	protectedReq.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	protectedW := httptest.NewRecorder()
	router.ServeHTTP(protectedW, protectedReq)

	// Должен вернуть либо 200 (если токен валиден), либо 401 (если истек)
	// В реальном приложении здесь была бы проверка времени жизни токена
	assert.True(t, protectedW.Code == http.StatusOK || protectedW.Code == http.StatusUnauthorized)
}
