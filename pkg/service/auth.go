package service

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/ChadaevArtem/rest-go-for-vue"
	"github.com/ChadaevArtem/rest-go-for-vue/pkg/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
)

const (
	salt                  = "asdagedrhftyki518sadf5as8"
	signingKey            = "awsg8s#@4Sf86DS#$2dF"
	accessTokenTTL        = time.Minute * 15
	refreshTokenTTL       = time.Hour * 24 * 365
	updateRefreshTokenTTL = time.Hour * 24 * 90
)

type tokenClaims struct {
	jwt.RegisteredClaims
	UserId int `json:"user_id"`
}

type AuthService struct {
	repo repository.Autorization
}

func NewAuthService(repo repository.Autorization) *AuthService {
	return &AuthService{repo: repo}
}

func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}

func generateRefreshToken() (string, error) {
	// 32 байта — это хорошая длина для безопасного токена.
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

func GenerateRefresh(userId int) (refresh rest.RefreshToken, err error) {
	// TODO: Сделать user-agent точнее название и описание девайса входа
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return
	}
	refresh = rest.RefreshToken{
		UserID:     userId,
		Token:      refreshToken,
		ExpiresAt:  time.Now().Add(refreshTokenTTL),
		NameDevice: "",
		DeviceInfo: "",
	}
	return
}

func (s *AuthService) CreateUser(user rest.User) (int, error) {
	user.Password = generatePasswordHash(user.Password)
	id, err := s.repo.CreateUser(user)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return 0, rest.ErrUserAlreadyExists
		}
	}
	return id, nil
}

func (s *AuthService) GenerateTokens(email, password string) (tokens rest.ResponseTokens, err error) {
	userId, err := s.repo.GetUser(email, generatePasswordHash(password))
	if err != nil {
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		userId,
	})

	accessToken, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return
	}

	refresh, err := GenerateRefresh(userId)
	if err != nil {
		return
	}

	if err = s.repo.CreateToken(refresh); err != nil {
		return
	}

	tokens = rest.ResponseTokens{
		AccessToken:  accessToken,
		RefreshToken: refresh.Token,
	}
	return
}

func (s *AuthService) GetAccessToken(refreshToken string) (tokens rest.ResponseTokens, err error) {
	userId, err := s.repo.GetUserIdByRefreshToken(refreshToken)
	if err != nil {
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		userId,
	})

	accessToken, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return
	}

	refresh, err := s.repo.GetRefreshToken(refreshToken)
	if err != nil {
		return
	}

	if refresh.ExpiresAt.Before(time.Now().Add(updateRefreshTokenTTL)) {
		refresh, err = GenerateRefresh(userId)
		err = s.repo.UpdateToken(refreshToken, refresh)
		if err != nil {
			return
		}
	}

	tokens = rest.ResponseTokens{
		AccessToken:  accessToken,
		RefreshToken: refresh.Token,
	}
	return
}

func (s *AuthService) ParseToken(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*tokenClaims)

	if !ok {
		return 0, errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserId, nil
}

func (s *AuthService) UnAuthorize(refreshToken string) error {
	refresh, err := s.repo.GetRefreshToken(refreshToken)
	if err != nil {
		return err
	}
	err = s.repo.DeleteRefreshToken(refresh.ID)

	return err
}

func (s *AuthService) UnAuthorizeAll(email, password string) error {
	id, err := s.repo.GetUser(email, password)
	if err != nil {
		return err
	}
	err = s.repo.DeleteAllUserRefreshTokens(id)
	return err
}
