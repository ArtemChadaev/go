package repository

import (
	"time"

	"github.com/ArtemChadaev/go"
	"github.com/jmoiron/sqlx"
)

type Autorization interface {
	CreateUser(user rest.User) (int, error)
	GetUser(username, password string) (int, error)
	GetUserEmailFromId(id int) (string, error)
	UpdateUserPassword(user rest.User) error
	GetUserIdByRefreshToken(refreshToken string) (int, error)
	CreateToken(refreshToken rest.RefreshToken) error
	GetRefreshToken(refreshToken string) (rest.RefreshToken, error)
	UpdateToken(oldRefreshToken string, refreshToken rest.RefreshToken) error
	DeleteRefreshToken(tokenId int) error
	DeleteAllUserRefreshTokens(userId int) error
	GetRefreshTokens(userId int) ([]rest.RefreshToken, error)
}
type UserSettings interface {
	CreateUserSettings(settings rest.UserSettings) error
	GetUserSettings(userId int) (rest.UserSettings, error)
	UpdateUserSettings(settings rest.UserSettings) error
	UpdateUserCoin(userId int, coin int) error
	BuyPaidSubscription(userId int, time time.Time) error
	DeactivateExpiredSubscriptions() (int64, error)
}

type Repository struct {
	Autorization
	UserSettings
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Autorization: NewAuthPostgres(db),
		UserSettings: NewUserSettingsPostgres(db),
	}
}
