package repository

import (
	"github.com/ArtemChadaev/go"
	"github.com/jmoiron/sqlx"
)

type AuthRepository struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(user rest.User) (int, error) {
	var id int
	query := "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id"
	row := r.db.QueryRow(query, user.Email, user.Password)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AuthRepository) GetUser(email, password string) (int, error) {
	var id int
	query := "SELECT id FROM users WHERE email=$1 and password_hash=$2"
	err := r.db.Get(&id, query, email, password)

	return id, err
}

func (r *AuthRepository) GetUserEmailFromId(id int) (string, error) {
	var userEmail string
	query := "SELECT email FROM users WHERE id=$1"
	err := r.db.Get(&userEmail, query, id)
	if err != nil {
		return "", err
	}
	return userEmail, err
}

func (r *AuthRepository) UpdateUserPassword(user rest.User) error {
	query := "UPDATE users SET password_hash=$1 WHERE id=$2"
	_, err := r.db.Exec(query, user.Password, user.ID)
	return err
}

func (r *AuthRepository) GetUserIdByRefreshToken(refreshToken string) (int, error) {
	var userId int
	query := "SELECT user_id FROM user_refresh_tokens WHERE token=$1"
	err := r.db.Get(&userId, query, refreshToken)
	if err != nil {
		return 0, err
	}
	return userId, err
}

func (r *AuthRepository) CreateToken(refreshToken rest.RefreshToken) error {
	query := "INSERT INTO user_refresh_tokens (user_id, token, expires_at, name_device, device_info) VALUES ($1, $2, $3, $4, $5)"
	_, err := r.db.Exec(query, refreshToken.UserID, refreshToken.Token, refreshToken.ExpiresAt, refreshToken.NameDevice, refreshToken.DeviceInfo)
	return err
}

func (r *AuthRepository) GetRefreshToken(refreshToken string) (rest.RefreshToken, error) {
	var refresh rest.RefreshToken
	query := "SELECT * FROM user_refresh_tokens WHERE token=$1"
	err := r.db.Get(&refresh, query, refreshToken)
	return refresh, err
}

func (r *AuthRepository) UpdateToken(oldRefreshToken string, refreshToken rest.RefreshToken) error {
	query := "UPDATE user_refresh_tokens SET token=$1, expires_at=$2, name_device=$3, device_info=$4 WHERE token=$5"
	_, err := r.db.Exec(query, refreshToken.Token, refreshToken.ExpiresAt, refreshToken.NameDevice, refreshToken.DeviceInfo, oldRefreshToken)
	return err
}

func (r *AuthRepository) DeleteRefreshToken(tokenId int) error {
	query := "DELETE FROM user_refresh_tokens WHERE id=$1"
	_, err := r.db.Exec(query, tokenId)
	return err
}

func (r *AuthRepository) DeleteAllUserRefreshTokens(userId int) error {
	query := "DELETE FROM user_refresh_tokens WHERE user_id=$1"
	_, err := r.db.Exec(query, userId)
	return err
}

func (r *AuthRepository) GetRefreshTokens(userId int) ([]rest.RefreshToken, error) {
	var refresh []rest.RefreshToken
	query := "SELECT * FROM user_refresh_tokens WHERE user_id=$1"
	err := r.db.Select(&refresh, query, userId)
	return refresh, err
}
