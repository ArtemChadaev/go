package unit

import (
	"database/sql"
	"testing"

	"github.com/ArtemChadaev/go/pkg/models"
	"github.com/ArtemChadaev/go/pkg/service"
	"github.com/ArtemChadaev/go/tests/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUserSettingsService_CreateInitialUserSettings_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageUserSettings(ctrl)
	
	// Создаем сервис без Redis для простых тестов
	userSettingsService := service.NewUserSettingsService(mockRepo, nil)

	userId := 1
	name := "testuser"

	mockRepo.EXPECT().CreateUserSettings(gomock.Any()).Return(nil)

	err := userSettingsService.CreateInitialUserSettings(userId, name)

	assert.NoError(t, err)
}

func TestUserSettingsService_GetByUserID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageUserSettings(ctrl)
	
	userSettingsService := service.NewUserSettingsService(mockRepo, nil)

	userId := 1
	expectedSettings := models.UserSettings{
		UserID: userId,
		Name:   "testuser",
		Coin:   100,
	}

	mockRepo.EXPECT().GetUserSettings(userId).Return(expectedSettings, nil)

	settings, err := userSettingsService.GetByUserID(userId)

	assert.NoError(t, err)
	assert.Equal(t, expectedSettings, settings)
}

func TestUserSettingsService_GetByUserID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageUserSettings(ctrl)
	
	userSettingsService := service.NewUserSettingsService(mockRepo, nil)

	userId := 1

	mockRepo.EXPECT().GetUserSettings(userId).Return(models.UserSettings{}, sql.ErrNoRows)

	settings, err := userSettingsService.GetByUserID(userId)

	assert.Error(t, err)
	assert.Equal(t, models.UserSettings{}, settings)
}

func TestUserSettingsService_UpdateInfo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageUserSettings(ctrl)
	
	userSettingsService := service.NewUserSettingsService(mockRepo, nil)

	userId := 1
	newName := "newname"
	newIcon := "icon.png"

	existingSettings := models.UserSettings{
		UserID: userId,
		Name:   "oldname",
		Coin:   100,
	}

	updatedSettings := models.UserSettings{
		UserID: userId,
		Name:   newName,
		Icon:   &newIcon,
		Coin:   100,
	}

	mockRepo.EXPECT().GetUserSettings(userId).Return(existingSettings, nil)
	mockRepo.EXPECT().UpdateUserSettings(updatedSettings).Return(nil)

	err := userSettingsService.UpdateInfo(userId, newName, newIcon)

	assert.NoError(t, err)
}

func TestUserSettingsService_UpdateInfo_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageUserSettings(ctrl)
	
	userSettingsService := service.NewUserSettingsService(mockRepo, nil)

	userId := 1
	newName := "newname"
	newIcon := "icon.png"

	mockRepo.EXPECT().GetUserSettings(userId).Return(models.UserSettings{}, sql.ErrNoRows)

	err := userSettingsService.UpdateInfo(userId, newName, newIcon)

	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestUserSettingsService_ChangeCoins_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageUserSettings(ctrl)
	
	userSettingsService := service.NewUserSettingsService(mockRepo, nil)

	userId := 1
	amount := 50
	currentCoins := 100
	newCoins := currentCoins + amount

	existingSettings := models.UserSettings{
		UserID: userId,
		Name:   "testuser",
		Coin:   currentCoins,
	}

	mockRepo.EXPECT().GetUserSettings(userId).Return(existingSettings, nil)
	mockRepo.EXPECT().UpdateUserCoin(userId, newCoins).Return(nil)

	err := userSettingsService.ChangeCoins(userId, amount)

	assert.NoError(t, err)
}

func TestUserSettingsService_ChangeCoins_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageUserSettings(ctrl)
	
	userSettingsService := service.NewUserSettingsService(mockRepo, nil)

	userId := 1
	amount := 50

	mockRepo.EXPECT().GetUserSettings(userId).Return(models.UserSettings{}, sql.ErrNoRows)

	err := userSettingsService.ChangeCoins(userId, amount)

	assert.Error(t, err)
	assert.Equal(t, models.ErrUserNotFound, err)
}

func TestUserSettingsService_ChangeCoins_InsufficientCoins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageUserSettings(ctrl)
	
	userSettingsService := service.NewUserSettingsService(mockRepo, nil)

	userId := 1
	amount := -150 // Пытаемся списать больше, чем есть
	currentCoins := 100

	existingSettings := models.UserSettings{
		UserID: userId,
		Name:   "testuser",
		Coin:   currentCoins,
	}

	mockRepo.EXPECT().GetUserSettings(userId).Return(existingSettings, nil)

	err := userSettingsService.ChangeCoins(userId, amount)

	assert.Error(t, err)
	assert.Equal(t, models.ErrNoCoins, err)
}

func TestUserSettingsService_ActivateSubscription_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageUserSettings(ctrl)
	
	userSettingsService := service.NewUserSettingsService(mockRepo, nil)

	userId := 1
	daysToAdd := 30
	paymentToken := "mock-success-payment-token"

	existingSettings := models.UserSettings{
		UserID:                 userId,
		Name:                   "testuser",
		Coin:                   100,
		PaidSubscription:        false,
		DateOfPaidSubscription:  nil,
	}

	mockRepo.EXPECT().GetUserSettings(userId).Return(existingSettings, nil)
	mockRepo.EXPECT().BuyPaidSubscription(userId, gomock.Any()).Return(nil)

	err := userSettingsService.ActivateSubscription(userId, daysToAdd, paymentToken)

	assert.NoError(t, err)
}

func TestUserSettingsService_ActivateSubscription_InvalidPaymentToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageUserSettings(ctrl)
	
	userSettingsService := service.NewUserSettingsService(mockRepo, nil)

	userId := 1
	daysToAdd := 30
	paymentToken := "invalid-token"

	err := userSettingsService.ActivateSubscription(userId, daysToAdd, paymentToken)

	assert.Error(t, err)
	assert.Equal(t, models.ErrPaymentFailed, err)
}

func TestUserSettingsService_ActivateSubscription_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockStorageUserSettings(ctrl)
	
	userSettingsService := service.NewUserSettingsService(mockRepo, nil)

	userId := 1
	daysToAdd := 30
	paymentToken := "mock-success-payment-token"

	mockRepo.EXPECT().GetUserSettings(userId).Return(models.UserSettings{}, sql.ErrNoRows)

	err := userSettingsService.ActivateSubscription(userId, daysToAdd, paymentToken)

	assert.Error(t, err)
	assert.Equal(t, models.ErrUserNotFound, err)
}
