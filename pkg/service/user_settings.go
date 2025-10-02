package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ArtemChadaev/go"
	"github.com/ArtemChadaev/go/pkg/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// Простой токен для имитации успешной оплаты в pet-проекте
	mockPaymentToken = "mock-success-payment-token"

	//Через каждые n минут удаляются все неактивные подписки
	checkExpiredSubscriptionsInterval = 10 * time.Minute

	//Количество монеток в день сколько можно получить
	dayCoins = 3
)

type UserSettingsService struct {
	repo  repository.UserSettings
	redis *redis.Client
}

func NewUserSettingsService(repo repository.UserSettings, redis *redis.Client) *UserSettingsService {
	service := &UserSettingsService{
		repo:  repo,
		redis: redis,
	}

	// Запускаем фоновую задачу для проверки подписок
	go service.startSubscriptionChecker()

	return service
}

// CreateInitialUserSettings создает начальные настройки для нового пользователя используется в auth.
func (s *UserSettingsService) CreateInitialUserSettings(userId int, name string) error {
	settings := rest.UserSettings{
		UserID:             userId,
		Name:               name, // Используется часть email до @
		DateOfRegistration: time.Now(),
	}
	return s.repo.CreateUserSettings(settings)
}

// GetByUserID возвращает настройки пользователя по его ID.
func (s *UserSettingsService) GetByUserID(userId int) (rest.UserSettings, error) {
	return s.repo.GetUserSettings(userId)
}

// UpdateInfo обновляет основную информацию пользователя (имя и иконку).
func (s *UserSettingsService) UpdateInfo(userId int, name, icon string) error {
	var settings rest.UserSettings
	settings, err := s.repo.GetUserSettings(userId)
	if err != nil {
		return err
	}
	settings.Name = name
	if icon != "" {
		settings.Icon = &icon
	}
	return s.repo.UpdateUserSettings(settings)
}

// ChangeCoins добавляет (или списывает) монеты пользователю.
func (s *UserSettingsService) ChangeCoins(userId, coin int) error {
	// Здесь бизнес-логика: мы не устанавливаем баланс, а изменяем его.
	settings, err := s.repo.GetUserSettings(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rest.ErrUserNotFound
		}
		return err
	}

	newBalance := settings.Coin + coin
	if newBalance < 0 {
		return rest.ErrNoCoins
	}

	return s.repo.UpdateUserCoin(userId, newBalance)
}

// ActivateSubscription активирует или продлевает подписку.
// В 'paymentToken' мы ожидаем некий токен от "платежной системы".
func (s *UserSettingsService) ActivateSubscription(userId, daysToAdd int, paymentToken string) error {
	// --- Защита от прямого вызова ---
	// В реальном проекте здесь была бы проверка токена через API платежной системы.
	// Для pet-проекта мы просто сравниваем его с константой.
	if paymentToken != mockPaymentToken {
		return rest.ErrPaymentFailed
	}

	settings, err := s.repo.GetUserSettings(userId)
	if err != nil {
		// Если настроек нет (хотя должны быть), возвращаем ошибку
		if errors.Is(err, sql.ErrNoRows) {
			return rest.ErrUserNotFound
		}
		return err
	}

	var newExpirationDate time.Time

	// --- Логика продления ---
	// Если подписка уже активна и не истекла, добавляем дни к дате окончания.
	if settings.PaidSubscription && settings.DateOfPaidSubscription != nil && settings.DateOfPaidSubscription.After(time.Now()) {
		newExpirationDate = settings.DateOfPaidSubscription.AddDate(0, 0, daysToAdd)
	} else {
		// Иначе, даем подписку от текущего момента.
		newExpirationDate = time.Now().AddDate(0, 0, daysToAdd)
	}

	return s.repo.BuyPaidSubscription(userId, newExpirationDate)
}

// GetGrantDailyReward даёт 3 монетки раз в день
func (s *UserSettingsService) GetGrantDailyReward(userId int) error {
	// Ключ в Redis будет уникальным для каждого дня, например, "daily_rewards:2025-09-27"
	key := "daily_rewards:" + time.Now().UTC().Format("2006-01-02")

	// SAdd добавит ID пользователя в "множество" (set) и вернет 1, если ID новый,
	// или 0, если ID там уже был. Это атомарная операция.
	added, err := s.redis.SAdd(context.Background(), key, userId).Result()
	if err != nil {
		return err
	}

	// Если ID не был добавлен (уже там был), значит награду получал.
	if added == 0 {
		return rest.ErrDayCoin // Награда уже получена сегодня
	}

	// Если ID был добавлен успешно, значит это первая выдача сегодня.
	// Устанавливаем "срок жизни" для ключа, чтобы он автоматически удалился через 25 часов.
	// Это избавляет нас от необходимости сбрасывать кеш вручную.
	s.redis.Expire(context.Background(), key, 24*time.Hour)

	// Теперь обновляем монеты в основной БД.
	if err := s.ChangeCoins(userId, dayCoins); err != nil { // Предполагаем, что этот метод прибавляет монеты
		s.redis.SRem(context.Background(), key, userId)
		return err
	}

	return nil
}

// startSubscriptionChecker - это внутренний метод, который в бесконечном цикле
// проверяет и деактивирует просроченные подписки.
func (s *UserSettingsService) startSubscriptionChecker() {
	ticker := time.NewTicker(checkExpiredSubscriptionsInterval)
	defer ticker.Stop()

	logrus.Infof("Запущена фоновая задача: проверка подписок каждые %v", checkExpiredSubscriptionsInterval)

	for {
		<-ticker.C // Ожидаем следующего тика

		logrus.Info("Выполняется плановая проверка просроченных подписок...")
		rowsAffected, err := s.repo.DeactivateExpiredSubscriptions()
		if err != nil {
			logrus.Errorf("Ошибка при деактивации просроченных подписок: %v", err)
			continue
		}

		if rowsAffected > 0 {
			logrus.Infof("Успешно деактивировано %d просроченных подписок", rowsAffected)
		}
	}
}
