package service

import (
	"github.com/ChadaevArtem/rest-go-for-vue"
	"github.com/ChadaevArtem/rest-go-for-vue/pkg/repository"
)

type Autorization interface {
	CreateUser(user rest.User) (int, error)
	GenerateTokens(email, password string) (tokens rest.ResponseTokens, err error)
	GetAccessToken(refreshToken string) (tokens rest.ResponseTokens, err error)
	ParseToken(accessToken string) (int, error)
	UnAuthorize(refreshToken string) error
	UnAuthorizeAll(email, password string) error
}

type Service struct {
	Autorization
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Autorization: NewAuthService(repos.Autorization),
	}
}
