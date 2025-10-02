package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ArtemChadaev/go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// OauthError остается как структура для JSON-ответа.
type OauthError struct {
	ErrorField       string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// handleError обрабатывает любую ошибку, пришедшую из слоев ниже.
func handleError(c *gin.Context, err error) {
	var appErr *rest.AppError

	// Проверяем, является ли ошибка нашей кастомной AppError.
	if errors.As(err, &appErr) {
		// Формируем сообщение для лога, включая исходную ошибку, если она есть.
		logMessage := appErr.Message
		if appErr.Err != nil {
			logMessage = fmt.Sprintf("%s: %v", appErr.Message, appErr.Err)
		}
		logrus.Error(logMessage)

		// Отправляем клиенту ответ со статусом и телом из нашей ошибки.
		c.AbortWithStatusJSON(appErr.HTTPStatus, OauthError{
			ErrorField:       appErr.Code,
			ErrorDescription: appErr.Message,
		})
	} else {
		// Если это неизвестная, непредвиденная ошибка.
		logrus.Errorf("unexpected error: %v", err)

		// Отправляем стандартный ответ о внутренней ошибке сервера.
		c.AbortWithStatusJSON(http.StatusInternalServerError, OauthError{
			ErrorField:       "internal_server_error",
			ErrorDescription: "An internal server error occurred.",
		})
	}
}
