package handler

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/ArtemChadaev/go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getUserID извлекает ID пользователя из контекста.
// Это вспомогательная функция, чтобы не дублировать код в каждом обработчике.
func getUserID(c *gin.Context) (int, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		return 0, rest.ErrInvalidToken
	}

	idInt, ok := id.(int)
	if !ok {
		return 0, rest.ErrInvalidToken
	}

	return idInt, nil
}

// getMySettings — пример обработчика для получения настроек текущего пользователя.
func (h *Handler) getMySettings(c *gin.Context) {
	// 1. Получаем ID пользователя из контекста с помощью нашей вспомогательной функции
	userId, err := getUserID(c)
	if err != nil {
		handleError(c, err)
		return
	}

	// 2. Используем полученный userId для вызова метода сервиса
	settings, err := h.services.GetByUserID(userId)
	if err != nil {
		handleError(c, err)
		return
	}

	// 3. Отправляем успешный ответ
	c.JSON(http.StatusOK, settings)
}

// setNameIcon Обновление имени и фота профиля
func (h *Handler) setNameIcon(c *gin.Context) {
	userId, err := getUserID(c)
	if err != nil {
		handleError(c, err)
		return
	}

	// 1. Получаем имя из поля формы
	// c.PostForm() извлечет значение поля "name" из multipart-формы
	newName := c.PostForm("name")
	if newName == "" {
		handleError(c, &rest.AppError{
			HTTPStatus: http.StatusBadRequest,
			Code:       "bad_request",
			Message:    "Поле 'name' не может быть пустым.",
		})
		return
	}

	iconUrl := "" // Переменная для хранения URL нового файла

	// 2. Проверяем, был ли загружен файл
	// c.FormFile() пытается получить файл из поля "icon"
	file, err := c.FormFile("icon")

	// Если err == nil, значит файл был прислан.
	if err == nil {
		// Генерируем уникальное имя файла
		ext := filepath.Ext(file.Filename)
		uniqueFilename := uuid.New().String() + ext

		// Сохраняем файл
		savePath := filepath.Join("static", "icons", uniqueFilename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			handleError(c, rest.ErrFailedSaveImg)
			return
		}
		// Формируем URL для сохранения в БД
		iconUrl = "/static/icons/" + uniqueFilename
	} else if errors.Is(err, http.ErrMissingFile) {
		// Если ошибка - это НЕ "файл отсутствует", значит произошла другая проблема.
		handleError(c, rest.NewInternalServerError(err))
		return
	}

	// 3. Вызываем сервис для обновления данных
	if err := h.services.UpdateInfo(userId, newName, iconUrl); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Профиль успешно обновлен",
		"iconUrl": iconUrl,
	})
}

// dayCoin Получить n монеток
func (h *Handler) dayCoin(c *gin.Context) {
	userId, err := getUserID(c)
	if err != nil {
		handleError(c, err)
		return
	}

	if err = h.services.GetGrantDailyReward(userId); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusConflict, gin.H{})
}

// setCoin пока что так, но надо система с генерацией ключей чтобы не обманывать систему
// func (h *Handler) setCoin(c *gin.Context) {}
