package health

import (
	"context"
	"net/http"

	"github.com/MrAndreID/goapi/v2/internal/application/cache"
	messageBroker "github.com/MrAndreID/goapi/v2/internal/application/message_broker"
	objectStorage "github.com/MrAndreID/goapi/v2/internal/application/object_storage"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository struct {
	Database      *gorm.DB
	Cache         *cache.CacheConnection
	MessageBroker *messageBroker.MessageBrokerConnection
	ObjectStorage *objectStorage.ObjectStorageConnection
}

func NewRepository(db *gorm.DB, cacheConnection *cache.CacheConnection, messageBrokerConnection *messageBroker.MessageBrokerConnection, objectStorageConnection *objectStorage.ObjectStorageConnection) *Repository {
	return &Repository{
		Database:      db,
		Cache:         cacheConnection,
		MessageBroker: messageBrokerConnection,
		ObjectStorage: objectStorageConnection,
	}
}

type InterfaceRepository interface {
	CheckDatabase(context.Context) (bool, error)
	CheckCache(context.Context) (bool, error)
	CheckMessageBroker(context.Context) (bool, error)
	CheckObjectStorage(context.Context) (bool, error)
}

func (r *Repository) CheckDatabase(ctx context.Context) (bool, error) {
	var tag string = "internal.feature.health.repository.CheckDatabase."

	if r.Database != nil {
		sqlDB, err := r.Database.DB()

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "01",
				"error": err.Error(),
			}).Error("failed to initialization sql database")

			return false, err
		}

		if err := sqlDB.PingContext(ctx); err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "02",
				"error": err.Error(),
			}).Error("failed to ping database")

			return false, err
		}
	}

	return true, nil
}

func (r *Repository) CheckCache(ctx context.Context) (bool, error) {
	var tag string = "internal.feature.health.repository.CheckCache."

	if r.Cache != nil {
		if r.Cache.Redis != nil {
			if err := r.Cache.Redis.Ping(ctx).Err(); err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "01",
					"error": err.Error(),
				}).Error("failed to ping redis cache")

				return false, err
			}
		}

		if r.Cache.Memcached != nil {
			if err := r.Cache.Memcached.Ping(); err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "02",
					"error": err.Error(),
				}).Error("failed to ping memcached")

				return false, err
			}
		}
	}

	return true, nil
}

func (r *Repository) CheckMessageBroker(ctx context.Context) (bool, error) {
	var tag string = "internal.feature.health.repository.CheckMessageBroker."

	if r.MessageBroker != nil {
		if r.MessageBroker.RabbitMQ != nil {
			if r.MessageBroker.RabbitMQ.Connection == nil || r.MessageBroker.RabbitMQ.Connection.IsClosed() {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "01",
					"error": "RabbitMQ Connection Is Closed",
				}).Error("failed to check rabbitmq connection")

				return false, nil
			}
		}

		if r.MessageBroker.Kafka != nil {
			if _, err := r.MessageBroker.Kafka.Controller(); err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "02",
					"error": err.Error(),
				}).Error("failed to reach kafka controller")

				return false, err
			}
		}
	}

	return true, nil
}

func (r *Repository) CheckObjectStorage(ctx context.Context) (bool, error) {
	var tag string = "internal.feature.health.repository.CheckObjectStorage."

	if r.ObjectStorage != nil {
		if r.ObjectStorage.Minio != nil {
			if _, err := r.ObjectStorage.Minio.ListBuckets(ctx); err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "01",
					"error": err.Error(),
				}).Error("failed to list minio buckets")

				return false, err
			}
		}

		if r.ObjectStorage.SeaweedFS != nil {
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, r.ObjectStorage.SeaweedFS.URL, nil)

			if err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "02",
					"error": err.Error(),
				}).Error("failed to build seaweedfs request")

				return false, err
			}

			response, err := http.DefaultClient.Do(request)

			if err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "03",
					"error": err.Error(),
				}).Error("failed to reach seaweedfs")

				return false, err
			}

			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "04",
					"error": "SeaweedFS Returned Non-200 Status",
				}).Error("failed to reach seaweedfs")

				return false, nil
			}
		}
	}

	return true, nil
}
