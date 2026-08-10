package application

import (
	"time"

	"github.com/MrAndreID/goapi/v2/internal/application/cache"
	"github.com/MrAndreID/goapi/v2/internal/application/config"
	"github.com/MrAndreID/goapi/v2/internal/application/database"
	messageBroker "github.com/MrAndreID/goapi/v2/internal/application/message_broker"
	objectStorage "github.com/MrAndreID/goapi/v2/internal/application/object_storage"
	userV1 "github.com/MrAndreID/goapi/v2/internal/feature/v1/user"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	UserV1Service *userV1.Service
)

func initService(app *Application) {
	UserV1Service = userV1.NewService(userV1.NewRepository(app.TimeLocation, app.Database))
}

func Start(toggle bool) any {
	var tag string = "internal.application.bootstrap.Start."

	cfg, err := config.New(toggle)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to initiate configuration")

		return nil
	}

	timeLocation, err := time.LoadLocation(cfg.AppLocation)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to load location for time")

		return nil
	}

	var databaseConnection *gorm.DB

	if cfg.UseDatabase {
		databaseConnection, err = database.New(&database.Database{
			Connection: cfg.DatabaseConnection,
			Host:       cfg.DatabaseHost,
			Port:       cfg.DatabasePort,
			Username:   cfg.DatabaseUsername,
			Password:   cfg.DatabasePassword,
			Name:       cfg.DatabaseName,
			SSLMode:    cfg.DatabaseSSLMode,
			ParseTime:  cfg.DatabaseParseTime,
			Charset:    cfg.DatabaseCharset,
			Timezone:   cfg.DatabaseTimezone,
		}, cfg.AppDebug)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "03",
				"error": err.Error(),
			}).Error("failed to connect database")

			return nil
		}
	}

	var cacheConnection *cache.CacheConnection

	if cfg.UseCache {
		cacheConnection, err = cache.New(&cache.Cache{
			Connection: cfg.CacheConnection,
			Host:       cfg.CacheHost,
			Port:       cfg.CachePort,
			Username:   cfg.CacheUsername,
			Password:   cfg.CachePassword,
		})

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "04",
				"error": err.Error(),
			}).Error("failed to connect cache")

			return nil
		}
	}

	var objectStorageConnection *objectStorage.ObjectStorageConnection

	if cfg.UseObjectStorage {
		objectStorageConnection, err = objectStorage.New(&objectStorage.ObjectStorage{
			Connection: cfg.ObjectStorageConnection,
			Host:       cfg.ObjectStorageHost,
			Port:       cfg.ObjectStoragePort,
			Username:   cfg.ObjectStorageUsername,
			Password:   cfg.ObjectStoragePassword,
			SSL:        cfg.ObjectStorageSSL,
		})

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "05",
				"error": err.Error(),
			}).Error("failed to connect object storage")

			return nil
		}
	}

	var messageBrokerConnection *messageBroker.MessageBrokerConnection

	if cfg.UseMessageBroker {
		messageBrokerConnection, err = messageBroker.New(&messageBroker.MessageBroker{
			Connection: cfg.MessageBrokerConnection,
			Host:       cfg.MessageBrokerHost,
			Port:       cfg.MessageBrokerPort,
			Username:   cfg.MessageBrokerUsername,
			Password:   cfg.MessageBrokerPassword,
			Name:       cfg.MessageBrokerName,
			Partition:  cfg.MessageBrokerPartition,
		})

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "06",
				"error": err.Error(),
			}).Error("failed to connect message broker")

			return nil
		}
	}

	if messageBrokerConnection != nil {
		defer messageBrokerConnection.Close()
	}

	app := &Application{
		Config:        cfg,
		TimeLocation:  timeLocation,
		Database:      databaseConnection,
		Cache:         cacheConnection,
		ObjectStorage: objectStorageConnection,
		MessageBroker: messageBrokerConnection,
	}

	initService(app)

	e := newServer(cfg)

	routes := RegisterRoutes(e, app)

	if toggle {
		return e.Start(":" + cfg.AppPort)
	}

	return routes
}
