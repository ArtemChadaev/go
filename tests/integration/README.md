# Интеграционные тесты

Эта директория содержит интеграционные тесты для Go-приложения, которые проверяют взаимодействие всех компонентов системы вместе.

## Структура

- `test_helper.go` - Вспомогательные функции для настройки тестового окружения
- `auth_test.go` - Тесты для аутентификационных эндпоинтов
- `user_settings_test.go` - Тесты для эндпоинтов настроек пользователя

## Требования

Для запуска интеграционных тестов требуются:

1. **Docker** - для запуска тестовых контейнеров PostgreSQL и Redis
2. **Go 1.25.1+** - для выполнения тестов
3. **Docker Desktop** (для Windows/macOS) или **Docker Engine** (для Linux)

## Установка зависимостей

```bash
# Установка зависимостей Go
go mod download

# Установка testcontainers (если еще не установлены)
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
go get github.com/testcontainers/testcontainers-go/modules/redis
```

## Запуск тестов

### Важно: Проблемы с Docker на Windows

TestContainers может не работать на Windows с rootless Docker. Если вы столкнулись с ошибкой `rootless Docker is not supported on Windows`, используйте локальные тесты:

```bash
# Локальные тесты (работают без Docker)
go test -v ./tests/integration/ -run TestAuthIntegrationLocal
go test -v ./tests/integration/ -run TestUserSettingsIntegrationLocal
```

### Запуск всех интеграционных тестов

```bash
# Из корня проекта
go test ./tests/integration/... -v

# Или из директории tests/integration
go test -v ./...
```

### Запуск конкретного теста

```bash
# Запуск только тестов аутентификации (Docker версия)
go test -v ./tests/integration/ -run TestAuthIntegration

# Запуск только тестов аутентификации (локальная версия)
go test -v ./tests/integration/ -run TestAuthIntegrationLocal

# Запуск только тестов настроек пользователя
go test -v ./tests/integration/ -run TestUserSettingsIntegration

# Запуск конкретного под-теста
go test -v ./tests/integration/ -run TestAuthIntegration/SignUp_and_SignIn_Flow
```

### Переменные окружения для локальных тестов

Для локальных тестов можно настроить следующие переменные:

```bash
# PostgreSQL
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=postgres
export TEST_DB_NAME=test_db

# Redis (опционально, тесты работают и без Redis)
export TEST_REDIS_HOST=localhost
export TEST_REDIS_PORT=6379
```

### Запуск с покрытием кода

```bash
go test -v -cover ./tests/integration/...
```

### Запуск с детальным выводом

```bash
go test -v -count=1 ./tests/integration/...
```

## Архитектура тестов

### TestContainers

Тесты используют [TestContainers](https://golang.testcontainers.org/) для запуска реальных баз данных:

- **PostgreSQL 16-alpine** - основная база данных
- **Redis 7-alpine** - для кэширования и rate limiting

Контейнеры автоматически запускаются перед каждым тестом и останавливаются после его завершения.

### Изоляция тестов

Каждый тест:
1. Создает чистые контейнеры баз данных
2. Выполняет миграции схемы
3. Запускает тестируемый код
4. Очищает все данные после завершения
5. Останавливает контейнеры

### Покрываемые сценарии

#### Auth тесты (`auth_test.go`)

- ✅ Регистрация нового пользователя
- ✅ Попытка повторной регистрации
- ✅ Вход в систему
- ✅ Вход с неверными данными
- ✅ Обновление токена
- ✅ Обновление с невалидным токеном
- ✅ Rate limiting для auth эндпоинтов
- ✅ Обработка невалидных запросов
- ✅ Проверка истечения токена

#### User Settings тесты (`user_settings_test.go`)

- ✅ Получение настроек пользователя
- ✅ Обновление настроек (имя, иконка)
- ✅ Получение ежедневных монет
- ✅ Повторное получение монет (должно быть заблокировано)
- ✅ Управление подпиской
- ✅ Доступ без токена (должен быть запрещен)
- ✅ Доступ с невалидным токеном
- ✅ Rate limiting для защищенных эндпоинтов
- ✅ Обработка невалидных запросов
- ✅ Конкурентный доступ к настройкам

## Переменные окружения

Тесты используют следующие переменные окружения (опционально):

```bash
# Для отладки можно установить логирование
export GIN_MODE=debug

# Для изменения таймаутов контейнеров
export TESTCONTAINERS_TIMEOUT=60s
```

## Устранение проблем

### Проблема: "Docker not running"

**Решение:** Убедитесь, что Docker запущен и доступен

```bash
# Проверка статуса Docker
docker info
```

### Проблема: "Port already in use"

**Решение:** Тесты автоматически выбирают свободные порты. Если проблема повторяется:

```bash
# Очистка неиспользуемых контейнеров
docker container prune -f

# Очистка неиспользуемых сетей
docker network prune -f
```

### Проблема: "Timeout waiting for container"

**Решение:** Увеличьте таймаут или проверьте ресурсы системы

```bash
# Увеличение таймаута
export TESTCONTAINERS_TIMEOUT=120s

# Проверка использования ресурсов
docker stats
```

### Проблема: "Permission denied"

**Решение:** Убедитесь, что пользователь имеет права на использование Docker

```bash
# Linux/macOS
sudo usermod -aG docker $USER
# После этого требуется перелогиниться

# Windows: запустить PowerShell от имени администратора
```

### Проблема: "Module not found"

**Решение:** Убедитесь, что вы в корневой директории проекта

```bash
# Проверка текущей директории
pwd
# Должно быть: /path/to/go

# Проверка структуры
ls -la
go.mod  # Должен существовать
```

## Производительность

- Время запуска одного теста: ~10-15 секунд (включая запуск контейнеров)
- Использование памяти: ~500MB для всех тестов
- Дисковое пространство: ~1GB для Docker образов

## CI/CD интеграция

Для GitHub Actions:

```yaml
name: Integration Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:24.0.5
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.25.1'
      - name: Run integration tests
        run: go test -v ./tests/integration/...
```

## Развитие тестов

При добавлении новых тестов:

1. Следуйте существующей структуре
2. Используйте `SetupTestSuite()` для инициализации
3. Очищайте данные с помощью `CleanupTestData()`
4. Добавляйте тесты в соответствующие файлы
5. Покрывайте как позитивные, так и негативные сценарии

### Пример добавления нового теста

```go
t.Run("New Feature Test", func(t *testing.T) {
    // Очистка данных
    err := testSuite.CleanupTestData(ctx)
    require.NoError(t, err)
    
    // Подготовка данных
    // ...
    
    // Выполнение тестируемого действия
    // ...
    
    // Проверка результатов
    assert.Equal(t, expected, actual)
})
```

## Полезные команды

```bash
# Просмотр доступных тестов
go test -list ./tests/integration/...

# Запуск с race condition детектором
go test -race -v ./tests/integration/...

# Запуск с бенчмарками
go test -bench=. -v ./tests/integration/...

# Генерация отчета о покрытии
go test -coverprofile=coverage.out -v ./tests/integration/...
go tool cover -html=coverage.out
