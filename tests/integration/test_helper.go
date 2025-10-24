package integration

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ArtemChadaev/go/pkg/api"
	"github.com/ArtemChadaev/go/pkg/service"
	"github.com/ArtemChadaev/go/pkg/storage"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	redislib "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestSuite содержит все зависимости для интеграционных тестов
type TestSuite struct {
	PostgresContainer *postgres.PostgresContainer
	RedisContainer    *rediscontainer.RedisContainer
	DB                *sqlx.DB
	RedisClient       *redislib.Client
	Handler           *api.Handler
	Cleanup           func()
}

// SetupTestSuite настраивает тестовое окружение с реальными базами данных
func SetupTestSuite(ctx context.Context) (*TestSuite, error) {
	// Загружаем .env файл если существует
	if err := godotenv.Load("../../.env"); err != nil {
		logrus.Warn("Не удалось загрузить .env файл")
	}

	// Настраиваем тестовую конфигурацию
	viper.Set("db.host", "localhost")
	viper.Set("db.port", "5432")
	viper.Set("db.username", "postgres")
	viper.Set("db.database", "test_db")
	viper.Set("db.sslmode", "disable")
	viper.Set("redis.addr", "localhost:6379")
	viper.Set("redis.db", 1)

	// Запускаем PostgreSQL контейнер
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось запустить PostgreSQL контейнер: %w", err)
	}

	// Получаем connection string для PostgreSQL
	dbHost, err := postgresContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить хост PostgreSQL: %w", err)
	}

	dbPort, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("не удалось получить порт PostgreSQL: %w", err)
	}

	// Запускаем Redis контейнер
	redisContainer, err := rediscontainer.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось запустить Redis контейнер: %w", err)
	}

	// Получаем connection string для Redis
	redisHost, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить хост Redis: %w", err)
	}

	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, fmt.Errorf("не удалось получить порт Redis: %w", err)
	}

	// Подключаемся к PostgreSQL
	db, err := storage.NewPostgresDB(storage.PostgresConfig{
		Host:     dbHost,
		Port:     dbPort.Port(),
		Username: "postgres",
		Database: "test_db",
		SSLMode:  "disable",
		Password: "postgres",
	})
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к PostgreSQL: %w", err)
	}

	// Подключаемся к Redis
	redisClient := redislib.NewClient(&redislib.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort.Port()),
		DB:   1,
	})

	// Проверяем подключение к Redis
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к Redis: %w", err)
	}

	// Запускаем миграции
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("не удалось запустить миграции: %w", err)
	}

	// Создаем зависимости
	repos := storage.NewRepository(db)
	services := service.NewService(repos, redisClient)
	handler := api.NewHandler(services, redisClient)

	// Функция очистки
	cleanup := func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			logrus.Errorf("Ошибка при остановке PostgreSQL контейнера: %v", err)
		}
		if err := redisContainer.Terminate(ctx); err != nil {
			logrus.Errorf("Ошибка при остановке Redis контейнера: %v", err)
		}
		if err := db.Close(); err != nil {
			logrus.Errorf("Ошибка при закрытии соединения с БД: %v", err)
		}
		if err := redisClient.Close(); err != nil {
			logrus.Errorf("Ошибка при закрытии соединения с Redis: %v", err)
		}
	}

	return &TestSuite{
		PostgresContainer: postgresContainer,
		RedisContainer:    redisContainer,
		DB:                db,
		RedisClient:       redisClient,
		Handler:           handler,
		Cleanup:           cleanup,
	}, nil
}

// runMigrations выполняет миграции базы данных
func runMigrations(db *sqlx.DB) error {
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
		icon_url VARCHAR(500),
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

// CleanupTestData очищает тестовые данные после каждого теста
func (ts *TestSuite) CleanupTestData(ctx context.Context) error {
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

	// Очищаем Redis
	if err := ts.RedisClient.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("ошибка при очистке Redis: %w", err)
	}

	return nil
}

// GetTestEnv возвращает переменные окружения для тестов
func GetTestEnv() map[string]string {
	return map[string]string{
		"DB_PASSWORD":    "postgres",
		"REDIS_PASSWORD": "",
	}
}

// SetTestEnv устанавливает переменные окружения для тестов
func SetTestEnv() {
	env := GetTestEnv()
	for key, value := range env {
		_ = os.Setenv(key, value)
	}
}
