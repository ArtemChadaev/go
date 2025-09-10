package service

import (
	"github.com/ChadaevArtem/rest-go-for-vue"
	"github.com/ChadaevArtem/rest-go-for-vue/pkg/repository"
)

type Autorization interface {
	CreateUser(user rest.User) (int, error)
	GenerateToken(email, password string) (string, error)
	ParseToken(token string) (int, error)
}

type Service struct {
	Autorization
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Autorization: NewAuthService(repos.Autorization),
	}
}
