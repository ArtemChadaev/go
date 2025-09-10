package repository

import (
	"github.com/ChadaevArtem/rest-go-for-vue"
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
	DeleteRefreshToken(token_id int) error
	DeleteAllUserRefreshTokens(userId int) error
	GetRefreshTokens(user_id int) ([]rest.RefreshToken, error)
}

type Repository struct {
	Autorization
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Autorization: NewAuthPostgres(db),
	}
}
