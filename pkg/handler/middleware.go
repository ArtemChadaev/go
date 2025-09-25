package handler

import (
	"strings"

	"github.com/ChadaevArtem/rest-go-for-vue"
	"github.com/gin-gonic/gin"
)

const (
	autorizationHeader = "Authorization"
	userCtx            = "userId"
)

// TODO: Проверка access токена посмотреть мб переделать
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

	userId, err := h.services.Autorization.ParseToken(headerParts[1])
	if err != nil {
		handleError(c, err) // Сервис уже вернет правильный rest.ErrInvalidToken
		return
	}

	c.Set(userCtx, userId)
}
