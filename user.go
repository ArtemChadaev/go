package rest

import "time"

type ResponseTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
type User struct {
	Id       int    `json:"-" db:"id"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshToken struct {
	ID         int       `db:"id"`
	UserID     int       `db:"user_id"`
	Token      string    `db:"token"`
	ExpiresAt  time.Time `db:"expires_at"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	NameDevice string    `db:"name_device"`
	DeviceInfo string    `db:"device_info"`
}
