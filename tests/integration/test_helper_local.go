package integration

import (
	"context"
	"fmt"
	"os"

	"github.com/ArtemChadaev/go/pkg/api"
	"github.com/ArtemChadaev/go/pkg/service"
	"github.com/ArtemChadaev/go/pkg/storage"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	redislib "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// TestSuiteLocal содержит все зависимости для интеграционных тестов (локальная версия)
type TestSuiteLocal struct {
	DB                *sqlx.DB
	RedisClient       *redislib.Client
	Handler           *api.Handler
	Cleanup           func()
}

// SetupTestSuiteLocal настраивает тестовое окружение с локальными базами данных
func SetupTestSuiteLocal(ctx context.Context) (*TestSuiteLocal, error) {
	// Загружаем .env файл если существует
	if err := godotenv.Load("../../.env"); err != nil {
		logrus.Warn("Не удалось загрузить .env файл")
	}

	// Используем переменные окружения или значения по умолчанию
	dbHost := getEnvOrDefault("TEST_DB_HOST", "localhost")
	dbPort := getEnvOrDefault("TEST_DB_PORT", "5432")
	dbUser := getEnvOrDefault("TEST_DB_USER", "postgres")
	dbPassword := getEnvOrDefault("TEST_DB_PASSWORD", "postgres")
	dbName := getEnvOrDefault("TEST_DB_NAME", "test_db")
	
	redisHost := getEnvOrDefault("TEST_REDIS_HOST", "localhost")
	redisPort := getEnvOrDefault("TEST_REDIS_PORT", "6379")

	// Подключаемся к PostgreSQL
	db, err := storage.NewPostgresDB(storage.PostgresConfig{
		Host:     dbHost,
		Port:     dbPort,
		Username: dbUser,
		Database: dbName,
		SSLMode:  "disable",
		Password: dbPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к PostgreSQL: %w", err)
	}

	// Подключаемся к Redis
	redisClient := redislib.NewClient(&redislib.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
		DB:   1,
	})

	// Проверяем подключение к Redis
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logrus.Warnf("Не удалось подключиться к Redis: %v. Тесты с Redis будут пропущены", err)
		// Не возвращаем ошибку, так как Redis может быть недоступен
	} else {
		logrus.Info("Подключение к Redis установлено")
	}

	// Запускаем миграции
	if err := runMigrationsLocal(db); err != nil {
		return nil, fmt.Errorf("не удалось запустить миграции: %w", err)
	}

	// Создаем зависимости
	repos := storage.NewRepository(db)
	services := service.NewService(repos, redisClient)
	handler := api.NewHandler(services, redisClient)

	// Функция очистки
	cleanup := func() {
		if err := db.Close(); err != nil {
			logrus.Errorf("Ошибка при закрытии соединения с БД: %v", err)
		}
		if err := redisClient.Close(); err != nil {
			logrus.Errorf("Ошибка при закрытии соединения с Redis: %v", err)
		}
	}

	return &TestSuiteLocal{
		DB:          db,
		RedisClient:  redisClient,
		Handler:      handler,
		Cleanup:      cleanup,
	}, nil
}

// runMigrationsLocal выполняет миграции базы данных
func runMigrationsLocal(db *sqlx.DB) error {
	// Читаем миграцию из файла
	migrationSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS user_settings (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		name VARCHAR(255),
		icon VARCHAR(500),
		subscription_expires_at TIMESTAMP,
		coins INTEGER DEFAULT 0,
		last_coin_claim TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id)
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);
	`

	_, err := db.Exec(migrationSQL)
	return err
}

// CleanupTestDataLocal очищает тестовые данные после каждого теста
func (ts *TestSuiteLocal) CleanupTestData(ctx context.Context) error {
	// Очищаем все таблицы в правильном порядке
	queries := []string{
		"DELETE FROM user_settings",
		"DELETE FROM users",
	}

	for _, query := range queries {
		if _, err := ts.DB.Exec(query); err != nil {
			return fmt.Errorf("ошибка при выполнении запроса '%s': %w", query, err)
		}
	}

	// Очищаем Redis если доступен
	if ts.RedisClient != nil {
		if err := ts.RedisClient.FlushDB(ctx).Err(); err != nil {
			logrus.Warnf("Ошибка при очистке Redis: %v", err)
		}
	}

	return nil
}

// getEnvOrDefault получает переменную окружения или возвращает значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDockerAvailable проверяет доступность Docker
func IsDockerAvailable() bool {
	// Простая проверка - можно расширить для более детальной проверки
	return os.Getenv("CI") != "" || os.Getenv("DOCKER_AVAILABLE") == "true"
}
