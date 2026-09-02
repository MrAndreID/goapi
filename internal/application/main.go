package application

import (
	"time"

	"github.com/MrAndreID/goapi/v2/internal/application/cache"
	"github.com/MrAndreID/goapi/v2/internal/application/config"
	messageBroker "github.com/MrAndreID/goapi/v2/internal/application/message_broker"
	objectStorage "github.com/MrAndreID/goapi/v2/internal/application/object_storage"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Application struct {
	Config        *config.Config
	TimeLocation  *time.Location
	Database      *gorm.DB
	Cache         *cache.CacheConnection
	ObjectStorage *objectStorage.ObjectStorageConnection
	MessageBroker *messageBroker.MessageBrokerConnection
}

func (app *Application) Close() {
	var tag string = "internal.application.main.Close."

	if app.MessageBroker != nil {
		app.MessageBroker.Close()
	}

	if app.Cache != nil {
		if app.Cache.Redis != nil {
			if err := app.Cache.Redis.Close(); err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "01",
					"error": err.Error(),
				}).Error("failed to close redis cache")
			}
		}

		if app.Cache.Memcached != nil {
			if err := app.Cache.Memcached.Close(); err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "02",
					"error": err.Error(),
				}).Error("failed to close memcached")
			}
		}
	}

	if app.Database != nil {
		databaseConnection, err := app.Database.DB()

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "03",
				"error": err.Error(),
			}).Error("failed to get the underlying database connection")

			return
		}

		if err := databaseConnection.Close(); err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "04",
				"error": err.Error(),
			}).Error("failed to close database")
		}
	}
}
