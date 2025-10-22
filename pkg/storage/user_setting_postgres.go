package repository

import (
	"time"

	"github.com/ArtemChadaev/go"
	"github.com/jmoiron/sqlx"
)

type UserSettingsRepository struct {
	db *sqlx.DB
}

func NewUserSettingsPostgres(db *sqlx.DB) *UserSettingsRepository {
	return &UserSettingsRepository{db: db}
}

func (r *UserSettingsRepository) CreateUserSettings(settings rest.UserSettings) error {
	query := "INSERT INTO user_settings (user_id, name) VALUES ($1, $2)"
	_, err := r.db.Exec(query, settings.UserID, settings.Name)
	return err
}

func (r *UserSettingsRepository) GetUserSettings(userId int) (rest.UserSettings, error) {
	var settings rest.UserSettings
	query := "SELECT * FROM user_settings WHERE user_id=$1"
	err := r.db.Get(&settings, query, userId)
	return settings, err
}

func (r *UserSettingsRepository) UpdateUserSettings(settings rest.UserSettings) error {
	query := "UPDATE user_settings SET name=$1, icon=$2 WHERE user_id=$3"
	_, err := r.db.Exec(query, settings.Name, settings.Icon, settings.UserID)
	return err
}

func (r *UserSettingsRepository) UpdateUserCoin(userId int, coin int) error {
	query := "UPDATE user_settings SET coin=$1 WHERE user_id=$2"
	_, err := r.db.Exec(query, coin, userId)
	return err
}
func (r *UserSettingsRepository) BuyPaidSubscription(userId int, time time.Time) error {
	query := "UPDATE user_settings SET paid_subscription=$1, date_of_paid_subscription=$2 WHERE user_id=$3"
	_, err := r.db.Exec(query, true, time, userId)
	return err
}
func (r *UserSettingsRepository) DeactivateExpiredSubscriptions() (int64, error) {
	query := `UPDATE user_settings SET paid_subscription = false 
			  WHERE paid_subscription = true AND date_of_paid_subscription < NOW()`

	result, err := r.db.Exec(query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
