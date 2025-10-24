package storage

import (
	"time"

	"github.com/ArtemChadaev/go/pkg/models"
	"github.com/jmoiron/sqlx"
)

type Autorization interface {
	CreateUser(user models.User) (int, error)
	GetUser(username, password string) (int, error)
	GetUserEmailFromId(id int) (string, error)
	UpdateUserPassword(user models.User) error
	GetUserIdByRefreshToken(refreshToken string) (int, error)
	CreateToken(refreshToken models.RefreshToken) error
	GetRefreshToken(refreshToken string) (models.RefreshToken, error)
	UpdateToken(oldRefreshToken string, refreshToken models.RefreshToken) error
	DeleteRefreshToken(tokenId int) error
	DeleteAllUserRefreshTokens(userId int) error
	GetRefreshTokens(userId int) ([]models.RefreshToken, error)
}
type UserSettings interface {
	CreateUserSettings(settings models.UserSettings) error
	GetUserSettings(userId int) (models.UserSettings, error)
	UpdateUserSettings(settings models.UserSettings) error
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
