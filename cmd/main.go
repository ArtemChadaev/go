package main

import (
	"os"

	"github.com/ArtemChadaev/go"
	"github.com/ArtemChadaev/go/pkg/handler"
	"github.com/ArtemChadaev/go/pkg/repository"
	"github.com/ArtemChadaev/go/pkg/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	if err := initConfig(); err != nil {
		logrus.Fatalf("%s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("%s", err.Error())
	}
	db, err := repository.NewPostgresDB(repository.PostgresConfig{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Database: viper.GetString("db.database"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("DB_PASSWORD"),
	})
	if err != nil {
		logrus.Fatalf("%s", err.Error())
	}
	redis, err := repository.NewRedisClient(repository.RedisConfig{
		Addr:     viper.GetString("redis.addr"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       viper.GetInt("redis.db"),
	})
	if err != nil {
		logrus.Fatalf("%s", err.Error())
	}
	repos := repository.NewRepository(db)
	services := service.NewService(repos, redis)
	handlers := handler.NewHandler(services, redis)

	srv := new(rest.Server)
	if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
		logrus.Fatalf("error http: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
