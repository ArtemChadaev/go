package handler

import (
	"context"
	"strings"
	"time"

	"github.com/ArtemChadaev/go"
	"github.com/gin-gonic/gin"
)

const (
	autorizationHeader = "Authorization"
	userCtx            = "userId"

	rateLimitPerMinute = 20
	rateWindow         = 1 * time.Minute

	authRateLimitPerMinute = 10
	authRateWindow         = 1 * time.Minute
)

// TODO: Проверка access токена посмотреть мб переделать

// Идентификация, проверка валидности токена только
func (h *Handler) userIdentify(c *gin.Context) {
	header := c.GetHeader(autorizationHeader)
	if header == "" {
		handleError(c, rest.ErrInvalidToken)
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		handleError(c, rest.ErrInvalidToken)
		return
	}

	userId, err := h.services.ParseToken(headerParts[1])
	if err != nil {
		handleError(c, err) // Сервис уже вернет правильный rest.ErrInvalidToken
		return
	}

	c.Set(userCtx, userId)
}

// rateLimiter - это middleware для ограничения частоты запросов по access токену
func (h *Handler) rateLimiter(c *gin.Context) {
	// 1. Извлекаем токен
	header := c.GetHeader(autorizationHeader)
	if header == "" {
		c.Next()
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		c.Next()
		return
	}
	accessToken := headerParts[1]

	// 2. Используем Redis для подсчета запросов
	ctx := context.Background()
	key := "rate_limit:" + accessToken

	// 3. Выполняем INCR и EXPIRE в одной транзакции (pipeline) для атомарности
	var count int64
	pipe := h.redis.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, rateWindow)
	_, err := pipe.Exec(ctx)
	if err != nil {
		// Если Redis недоступен, лучше пропустить запрос, чем заблокировать всех.
		// Можно добавить логирование.
		c.Next()
		return
	}

	count = incr.Val()

	// 4. Проверяем лимит
	if count > rateLimitPerMinute {
		handleError(c, rest.ErrTooManyRequestsByAccessToken)
		c.Abort() // Важно остановить дальнейшую обработку
		return
	}

	c.Next()
}

// authRateLimiter - это middleware для ограничения частоты запросов к эндпоинтам /auth по IP-адресу
func (h *Handler) authRateLimiter(c *gin.Context) {
	// 1. В качестве идентификатора используем IP-адрес клиента
	ip := c.ClientIP()
	key := "rate_limit_auth:" + ip
	ctx := context.Background()

	// 2. Используем тот же надежный алгоритм с Redis
	var count int64
	pipe := h.redis.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, authRateWindow)
	_, err := pipe.Exec(ctx)

	if err != nil {
		// Если Redis недоступен, пропускаем запрос, чтобы не блокировать весь сервис.
		// Здесь можно добавить логирование ошибки.
		c.Next()
		return
	}
	count = incr.Val()

	// 3. Проверяем лимит
	if count > authRateLimitPerMinute {
		handleError(c, rest.ErrTooManyRequestsByIp)
		c.Abort()
		return
	}

	c.Next()
}
