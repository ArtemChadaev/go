package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ArtemChadaev/go/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSettingsIntegration(t *testing.T) {
	ctx := context.Background()
	
	// Настраиваем тестовое окружение
	testSuite, err := SetupTestSuite(ctx)
	require.NoError(t, err)
	defer testSuite.Cleanup()

	// Устанавливаем gin в тестовом режиме
	gin.SetMode(gin.TestMode)

	// Создаем тестового пользователя и получаем токены
	user := models.User{
		Email:    "settings@example.com",
		Password: "password123",
	}

	// Регистрируем пользователя
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := testSuite.Handler.InitRoutes()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var tokens models.ResponseTokens
	err = json.Unmarshal(w.Body.Bytes(), &tokens)
	require.NoError(t, err)

	t.Run("Get User Settings", func(t *testing.T) {
		// Очищаем данные перед тестом
		err := testSuite.CleanupTestData(ctx)
		require.NoError(t, err)

		// Перерегистрируем пользователя
		body, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := testSuite.Handler.InitRoutes()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var newTokens models.ResponseTokens
		err = json.Unmarshal(w.Body.Bytes(), &newTokens)
		require.NoError(t, err)

		// Получаем настройки пользователя
		req = httptest.NewRequest("GET", "/api/settings/", nil)
		req.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var settings models.UserSettings
		err = json.Unmarshal(w.Body.Bytes(), &settings)
		assert.NoError(t, err)
		assert.Equal(t, 0, settings.Coin) // Начальное количество монет
		assert.Empty(t, settings.Name)     // Имя должно быть пустым
		assert.Nil(t, settings.Icon)       // Иконка должна быть пустой
	})

	t.Run("Update User Settings", func(t *testing.T) {
		// Очищаем данные перед тестом
		err := testSuite.CleanupTestData(ctx)
		require.NoError(t, err)

		// Перерегистрируем пользователя
		body, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := testSuite.Handler.InitRoutes()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var newTokens models.ResponseTokens
		err = json.Unmarshal(w.Body.Bytes(), &newTokens)
		require.NoError(t, err)

		// Создаем структуру для обновления настроек
		type UpdateUserSettings struct {
			Name  string `json:"name"`
			Icon  string `json:"icon"`
		}

		updateRequest := UpdateUserSettings{
			Name: "Test User",
			Icon: "https://example.com/icon.png",
		}

		body, _ = json.Marshal(updateRequest)
		req = httptest.NewRequest("PUT", "/api/settings/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Проверяем, что настройки обновились
		req = httptest.NewRequest("GET", "/api/settings/", nil)
		req.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var settings models.UserSettings
		err = json.Unmarshal(w.Body.Bytes(), &settings)
		assert.NoError(t, err)
		assert.Equal(t, "Test User", settings.Name)
		assert.Equal(t, "https://example.com/icon.png", *settings.Icon)
	})

	t.Run("Day Coin Claim", func(t *testing.T) {
		// Очищаем данные перед тестом
		err := testSuite.CleanupTestData(ctx)
		require.NoError(t, err)

		// Перерегистрируем пользователя
		body, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := testSuite.Handler.InitRoutes()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var newTokens models.ResponseTokens
		err = json.Unmarshal(w.Body.Bytes(), &newTokens)
		require.NoError(t, err)

		// Получаем начальные настройки
		req = httptest.NewRequest("GET", "/api/settings/", nil)
		req.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var initialSettings models.UserSettings
		err = json.Unmarshal(w.Body.Bytes(), &initialSettings)
		require.NoError(t, err)

		// Запрашиваем ежедневные монеты
		req = httptest.NewRequest("POST", "/api/settings/dayCoin", nil)
		req.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Проверяем, что количество монет увеличилось
		req = httptest.NewRequest("GET", "/api/settings/", nil)
		req.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedSettings models.UserSettings
		err = json.Unmarshal(w.Body.Bytes(), &updatedSettings)
		assert.NoError(t, err)
		assert.Greater(t, updatedSettings.Coin, initialSettings.Coin)

		// Пытаемся получить монеты еще раз (должно быть запрещено)
		req = httptest.NewRequest("POST", "/api/settings/dayCoin", nil)
		req.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("Subscription", func(t *testing.T) {
		// Очищаем данные перед тестом
		err := testSuite.CleanupTestData(ctx)
		require.NoError(t, err)

		// Перерегистрируем пользователя
		body, _ := json.Marshal(user)
		req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router := testSuite.Handler.InitRoutes()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var newTokens models.ResponseTokens
		err = json.Unmarshal(w.Body.Bytes(), &newTokens)
		require.NoError(t, err)

		// Проверяем подписку (эндпоинт существует, но логика может быть реализована по-разному)
		req = httptest.NewRequest("POST", "/api/settings/subscript", nil)
		req.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Статус может быть разным в зависимости от реализации
		// Проверяем, что запрос обработан (не 404)
		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		testCases := []struct {
			name     string
			method   string
			endpoint string
			body     interface{}
		}{
			{
				name:     "Get Settings Without Token",
				method:   "GET",
				endpoint: "/api/settings/",
				body:     nil,
			},
			{
				name:     "Update Settings Without Token",
				method:   "PUT",
				endpoint: "/api/settings/",
				body:     map[string]string{"name": "Test"},
			},
			{
				name:     "Day Coin Without Token",
				method:   "POST",
				endpoint: "/api/settings/dayCoin",
				body:     nil,
			},
			{
				name:     "Subscription Without Token",
				method:   "POST",
				endpoint: "/api/settings/subscript",
				body:     nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var reqBody []byte
				if tc.body != nil {
					reqBody, _ = json.Marshal(tc.body)
				}

				req := httptest.NewRequest(tc.method, tc.endpoint, bytes.NewBuffer(reqBody))
				req.Header.Set("Content-Type", "application/json")
				// Не устанавливаем заголовок авторизации

				w := httptest.NewRecorder()
				router := testSuite.Handler.InitRoutes()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
			})
		}
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/settings/", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")

		w := httptest.NewRecorder()
		router := testSuite.Handler.InitRoutes()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Requests", func(t *testing.T) {
		testCases := []struct {
			name           string
			method         string
			endpoint       string
			body           interface{}
			expectedStatus int
		}{
			{
				name:           "Update Settings Invalid JSON",
				method:         "PUT",
				endpoint:       "/api/settings/",
				body:           "{invalid json",
				expectedStatus: http.StatusBadRequest,
			},
			{
				name:           "Update Settings Empty Body",
				method:         "PUT",
				endpoint:       "/api/settings/",
				body:           nil,
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var reqBody []byte
				if str, ok := tc.body.(string); ok {
					reqBody = []byte(str)
				} else if tc.body != nil {
					reqBody, _ = json.Marshal(tc.body)
				}

				req := httptest.NewRequest(tc.method, tc.endpoint, bytes.NewBuffer(reqBody))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

				w := httptest.NewRecorder()
				router := testSuite.Handler.InitRoutes()
				router.ServeHTTP(w, req)

				assert.Equal(t, tc.expectedStatus, w.Code)
			})
		}
	})

	t.Run("Rate Limiting", func(t *testing.T) {
		// Проверяем rate limiting для защищенных эндпоинтов
		for i := 0; i < 15; i++ { // Превышаем лимит
			req := httptest.NewRequest("GET", "/api/settings/", nil)
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

			w := httptest.NewRecorder()
			router := testSuite.Handler.InitRoutes()
			router.ServeHTTP(w, req)

			if i < 10 { // Первые 10 запросов должны пройти
				assert.Equal(t, http.StatusOK, w.Code, "Request %d should pass", i)
			} else { // Последующие должны быть заблокированы
				assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request %d should be rate limited", i)
			}
		}
	})
}

func TestUserSettingsConcurrency(t *testing.T) {
	ctx := context.Background()
	
	// Настраиваем тестовое окружение
	testSuite, err := SetupTestSuite(ctx)
	require.NoError(t, err)
	defer testSuite.Cleanup()

	// Устанавливаем gin в тестовом режиме
	gin.SetMode(gin.TestMode)

	// Создаем тестового пользователя
	user := models.User{
		Email:    "concurrent@example.com",
		Password: "password123",
	}

	// Регистрируем пользователя
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := testSuite.Handler.InitRoutes()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var tokens models.ResponseTokens
	err = json.Unmarshal(w.Body.Bytes(), &tokens)
	require.NoError(t, err)

	// Тестируем конкурентный доступ к настройкам
	t.Run("Concurrent Settings Access", func(t *testing.T) {
		const numGoroutines = 10
		results := make(chan int, numGoroutines)

		// Запускаем несколько горутин для одновременного доступа
		for i := 0; i < numGoroutines; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/api/settings/", nil)
				req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

				w := httptest.NewRecorder()
				router := testSuite.Handler.InitRoutes()
				router.ServeHTTP(w, req)

				results <- w.Code
			}()
		}

		// Собираем результаты
		for i := 0; i < numGoroutines; i++ {
			statusCode := <-results
			// Все запросы должны быть успешными или заблокированными из-за rate limiting
			assert.True(t, statusCode == http.StatusOK || statusCode == http.StatusTooManyRequests)
		}
	})
}
