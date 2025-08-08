package caches

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Cache struct {
	Connection string
	Host       string
	Port       string
	Username   string
	Password   string
}

func New(cache *Cache) (*redis.Client, error) {
	var (
		cacheData *redis.Client
		err       error
	)

	switch cache.Connection {
	case "redis":
		cacheData, err = cache.Redis()
	default:
		err = errors.New("Cache Connection Not Found")
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "Caches.Main.New.01",
			"error": err.Error(),
		}).Error("failed to connect cache")

		return nil, err
	}

	return cacheData, nil
}

func (cache *Cache) Redis() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cache.Host + ":" + cache.Port,
		Username: cache.Username,
		Password: cache.Password,
		DB:       0,
	})

	_, err := redisClient.Ping(context.Background()).Result()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "Caches.Main.Redis.01",
			"error": err.Error(),
		}).Error("failed to connect redis cache")

		return nil, err
	}

	return redisClient, nil
}
