package rest

import (
	"net/http"
)

// AppError — это наша основная структура для всех ошибок приложения.
type AppError struct {
	// Статус-код, который вернется клиенту.
	HTTPStatus int `json:"-"`
	// Короткий код ошибки в стиле Oauth.
	Code string `json:"error"`
	// Описание.
	Message string `json:"error_description"`
	// Внутренняя (исходная) ошибка для логирования.
	Err error `json:"-"`
}

// Error позволяет AppError соответствовать стандартному интерфейсу error.
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap используется для извлечения исходной ошибки (стандарт Go 1.13+).
func (e *AppError) Unwrap() error {
	return e.Err
}

// Стандартные ошибки приложения
var (
	// email занят
	ErrUserAlreadyExists = &AppError{
		HTTPStatus: http.StatusConflict,
		Code:       "email_exist",
		Message:    "user with this email already exists",
	}
	// неверный email или пароль
	ErrInvalidCredentials = &AppError{
		HTTPStatus: http.StatusUnauthorized,
		Code:       "invalid_credentials",
		Message:    "invalid email or password",
	}
	// невалидный токен
	ErrInvalidToken = &AppError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "invalid_token",
		Message:    "authorization token is invalid",
	}
)

// Функции-конструкторы для ошибок, которые должны содержать дополнительный контекст.

// NewInvalidRequestError создает ошибку для некорректного запроса (например, невалидный JSON).
func NewInvalidRequestError(err error) *AppError {
	return &AppError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "invalid_request",
		Message:    "invalid request body or parameters",
		Err:        err,
	}
}

// NewInternalServerError создает ошибку для всех непредвиденных сбоев.
func NewInternalServerError(err error) *AppError {
	return &AppError{
		HTTPStatus: http.StatusInternalServerError,
		Code:       "internal_server_error",
		Message:    "an internal server error occurred",
		Err:        err,
	}
}