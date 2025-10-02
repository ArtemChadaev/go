package service

import (
	"github.com/ArtemChadaev/go"
	"github.com/ArtemChadaev/go/pkg/repository"
	"github.com/redis/go-redis/v9"
)

type Autorization interface {
	CreateUser(user rest.User) (int, error)
	GenerateTokens(email, password string) (tokens rest.ResponseTokens, err error)
	GetAccessToken(refreshToken string) (tokens rest.ResponseTokens, err error)
	ParseToken(accessToken string) (int, error)
	UnAuthorize(refreshToken string) error
	UnAuthorizeAll(email, password string) error
}
type UserSettings interface {
	CreateInitialUserSettings(userId int, name string) error
	GetByUserID(userId int) (rest.UserSettings, error)
	UpdateInfo(userId int, name, icon string) error
	ChangeCoins(userId, amount int) error
	ActivateSubscription(userId, daysToAdd int, paymentToken string) error
	GetGrantDailyReward(userId int) error
}
type Service struct {
	Autorization
	UserSettings
}

func NewService(repos *repository.Repository, redis *redis.Client) *Service {
	userSettingsService := NewUserSettingsService(repos.UserSettings, redis)

	authService := NewAuthService(repos.Autorization, userSettingsService)

	return &Service{
		Autorization: authService,
		UserSettings: userSettingsService,
	}
}
