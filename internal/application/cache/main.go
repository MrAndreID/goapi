package cache

import (
	"context"
	"errors"

	"github.com/bradfitz/gomemcache/memcache"
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

type CacheConnection struct {
	Redis     *redis.Client
	Memcached *memcache.Client
}

func New(cache *Cache) (*CacheConnection, error) {
	var (
		cacheData *CacheConnection
		err       error
	)

	switch cache.Connection {
	case "redis":
		cacheData, err = cache.Redis()
	case "memcached":
		cacheData, err = cache.Memcached()
	default:
		err = errors.New("Cache Connection Not Found")
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "internal.application.cache.main.New.01",
			"error": err.Error(),
		}).Error("failed to connect cache")

		return nil, err
	}

	return cacheData, nil
}

func (cache *Cache) Redis() (*CacheConnection, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cache.Host + ":" + cache.Port,
		Username: cache.Username,
		Password: cache.Password,
		DB:       0,
	})

	_, err := redisClient.Ping(context.Background()).Result()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "internal.application.cache.main.Redis.01",
			"error": err.Error(),
		}).Error("failed to connect redis cache")

		return nil, err
	}

	return &CacheConnection{
		Redis: redisClient,
	}, nil
}

func (cache *Cache) Memcached() (*CacheConnection, error) {
	mc := memcache.New(cache.Host + ":" + cache.Port)

	if err := mc.Ping(); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "internal.application.cache.main.Memcached.01",
			"error": err.Error(),
		}).Error("failed to connect memcached")

		return nil, err
	}

	return &CacheConnection{
		Memcached: mc,
	}, nil
}
