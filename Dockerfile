# Dockerfile

# --- ЭТАП 1: "Сборщик" ---
FROM golang:1.25.1-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
# Скачиваем зависимости ДО копирования кода.
# Этот шаг кэшируется и ускоряет сборку, если зависимости не менялись.
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь остальной исходный код
COPY . .

# Собираем СТАТИЧЕСКИЙ бинарный файл для Linux.
# CGO_ENABLED=0 - отключает CGO. Это КЛЮЧ к созданию автономного бинарника.
# -o myapp - имя нашего скомпилированного файла.
RUN CGO_ENABLED=0 GOOS=linux go build -a -o myapp ./cmd/main.go

# --- ЭТАП 2: "Финальный образ" ---
# Начинаем с НУЛЯ. `alpine` - один из самых маленьких (около 5MB)
# Он содержит только минимум (включая shell, что полезно для отладки).
FROM alpine:3.22

# Создаем пользователя "appuser" без привилегий
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем ТОЛЬКО бинарный файл из этапа "Сборщик"
COPY --from=builder /app/myapp .

# Копируем конфиги, если они нужны (хотя в K8s их лучше монтировать как ConfigMap)
COPY ./configs ./configs

# Даем права нашему пользователю на рабочую директорию
RUN chown -R appuser:appgroup /app

# Переключаемся на non-root пользователя
USER appuser

# Открываем порт, который слушает ваше приложение
EXPOSE 8080

# Команда запуска
CMD ["./myapp"]