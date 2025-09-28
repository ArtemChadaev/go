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

// Ошибки связанные с авторизацией
var (
	// ErrUserAlreadyExists email занят
	ErrUserAlreadyExists = &AppError{
		HTTPStatus: http.StatusConflict,
		Code:       "email_exist",
		Message:    "user with this email already exists",
	}
	// ErrInvalidCredentials неверный email или пароль
	ErrInvalidCredentials = &AppError{
		HTTPStatus: http.StatusUnauthorized,
		Code:       "invalid_credentials",
		Message:    "invalid email or password",
	}
	// ErrInvalidToken невалидный токен
	ErrInvalidToken = &AppError{
		HTTPStatus: http.StatusBadRequest,
		Code:       "invalid_token",
		Message:    "authorization token is invalid",
	}

	// ErrTooManyRequestsByAccessToken Превышено количество запросов по токену
	ErrTooManyRequestsByAccessToken = &AppError{
		HTTPStatus: http.StatusTooManyRequests,
		Code:       "too_many_requests",
		Message:    "too many requests by access token",
	}
	// ErrTooManyRequestsByIp Превышено количество запросов по ip
	ErrTooManyRequestsByIp = &AppError{
		HTTPStatus: http.StatusTooManyRequests,
		Code:       "too_many_requests",
		Message:    "too many requests by ip",
	}

	// ErrUserNotFound Пользователь не найден
	ErrUserNotFound = &AppError{
		HTTPStatus: http.StatusNotFound,
		Code:       "user_not_found",
		Message:    "user not found",
	}
)

// Ошибки связанные с настройкой
var (
	// ErrNoCoins Не хватает монеток на аккаунте
	ErrNoCoins = &AppError{
		HTTPStatus: http.StatusPaymentRequired,
		Code:       "no_coins",
		Message:    "there are not enough coins in the account",
	}
	// ErrFailedSaveImg Не удалось сохранить фотографию
	ErrFailedSaveImg = &AppError{
		HTTPStatus: http.StatusInternalServerError,
		Code:       "failed_save_img",
		Message:    "failed save img",
	}

	ErrDayCoin = &AppError{
		HTTPStatus: http.StatusConflict,
		Code:       "day_coin",
		Message:    "day coin",
	}
)

// Платёж всё связанное с ним
var (
	// ErrNoMoney Не хватает денег
	ErrNoMoney = &AppError{
		HTTPStatus: http.StatusPaymentRequired,
		Code:       "no_money",
		Message:    "there are not enough money in the account",
	}
	// ErrPaymentFailed Ошибка платежа
	ErrPaymentFailed = &AppError{
		HTTPStatus: http.StatusPaymentRequired,
		Code:       "payment_failed",
		Message:    "payment failed",
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
