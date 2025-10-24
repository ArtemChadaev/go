package service

import (
	"github.com/ArtemChadaev/go/pkg/models"
	"github.com/ArtemChadaev/go/pkg/storage"
	"github.com/redis/go-redis/v9"
)

type Autorization interface {
	CreateUser(user models.User) (int, error)
	GenerateTokens(email, password string) (tokens models.ResponseTokens, err error)
	GetAccessToken(refreshToken string) (tokens models.ResponseTokens, err error)
	ParseToken(accessToken string) (int, error)
	UnAuthorize(refreshToken string) error
	UnAuthorizeAll(email, password string) error
}
type UserSettings interface {
	CreateInitialUserSettings(userId int, name string) error
	GetByUserID(userId int) (models.UserSettings, error)
	UpdateInfo(userId int, name, icon string) error
	ChangeCoins(userId, amount int) error
	ActivateSubscription(userId, daysToAdd int, paymentToken string) error
	GetGrantDailyReward(userId int) error
}
type Service struct {
	Autorization
	UserSettings
}

func NewService(repos *storage.Repository, redis *redis.Client) *Service {
	userSettingsService := NewUserSettingsService(repos.UserSettings, redis)

	authService := NewAuthService(repos.Autorization, userSettingsService)

	return &Service{
		Autorization: authService,
		UserSettings: userSettingsService,
	}
}
