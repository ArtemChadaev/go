package service

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ArtemChadaev/go"
	"github.com/ArtemChadaev/go/pkg/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
)

const (
	//соль
	salt = "asdagedrhftyki518sadf5as8"
	// ключ
	signingKey = "awsg8s#@4Sf86DS#$2dF"
	//Время жизни accesToken
	accessTokenTTL = time.Minute * 15
	//Время жизни refreshToken
	refreshTokenTTL = time.Hour * 24 * 365
	//Время за которое refreshToken обновится
	updateRefreshTokenTTL = time.Hour * 24 * 90
)

type tokenClaims struct {
	jwt.RegisteredClaims
	UserId int `json:"user_id"`
}

type AuthService struct {
	repo            repository.Autorization
	settingsService UserSettings
}

func NewAuthService(repo repository.Autorization, settingsService UserSettings) *AuthService {
	return &AuthService{
		repo:            repo,
		settingsService: settingsService,
	}
}

func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}

func generateRefreshToken() (string, error) {
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
		NameDevice: nil,
		DeviceInfo: nil,
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
		return 0, rest.NewInternalServerError(err)
	}
	if err := s.settingsService.CreateInitialUserSettings(id, strings.Split(user.Email, "@")[0]); err != nil {
		return 0, rest.NewInternalServerError(err)
	}
	return id, nil
}

func (s *AuthService) GenerateTokens(email, password string) (tokens rest.ResponseTokens, err error) {
	userId, err := s.repo.GetUser(email, generatePasswordHash(password))
	if err != nil {
		// Если пользователь не найден - это ошибка неверных учетных данных.
		if errors.Is(err, sql.ErrNoRows) {
			return tokens, rest.ErrInvalidCredentials
		}
		// Иначе - внутренняя ошибка.
		return tokens, rest.NewInternalServerError(err)
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
		return tokens, rest.NewInternalServerError(err)
	}

	refresh, err := GenerateRefresh(userId)
	if err != nil {
		return tokens, rest.NewInternalServerError(err)
	}

	if err = s.repo.CreateToken(refresh); err != nil {
		return tokens, rest.NewInternalServerError(err)
	}

	tokens = rest.ResponseTokens{
		AccessToken:  accessToken,
		RefreshToken: refresh.Token,
	}
	return
}

func (s *AuthService) GetAccessToken(refreshToken string) (tokens rest.ResponseTokens, err error) {

	refresh, err := s.repo.GetRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tokens, rest.ErrInvalidToken
		}
		return tokens, rest.NewInternalServerError(err)
	}

	if time.Now().After(refresh.ExpiresAt) {
		_ = s.repo.DeleteRefreshToken(refresh.ID) // Удаляем "мусор" из БД
		return tokens, rest.ErrInvalidToken
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		refresh.UserID,
	})

	accessToken, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return tokens, rest.NewInternalServerError(err)
	}

	if refresh.ExpiresAt.Before(time.Now().Add(updateRefreshTokenTTL)) {
		newRefresh, err := GenerateRefresh(refresh.UserID)
		if err != nil {
			return tokens, rest.NewInternalServerError(err)
		}
		err = s.repo.UpdateToken(refreshToken, newRefresh)
		if err != nil {
			return tokens, rest.NewInternalServerError(err)
		}
		refresh.Token = newRefresh.Token
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
		return 0, rest.ErrInvalidToken
	}
	claims, ok := token.Claims.(*tokenClaims)

	if !ok {
		return 0, rest.ErrInvalidToken
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
