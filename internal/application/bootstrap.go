package application

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MrAndreID/goapi/v2/internal/application/cache"
	"github.com/MrAndreID/goapi/v2/internal/application/config"
	"github.com/MrAndreID/goapi/v2/internal/application/database"
	"github.com/MrAndreID/goapi/v2/internal/application/dependency"
	messageBroker "github.com/MrAndreID/goapi/v2/internal/application/message_broker"
	objectStorage "github.com/MrAndreID/goapi/v2/internal/application/object_storage"
	userV1 "github.com/MrAndreID/goapi/v2/internal/feature/v1/user"

	"github.com/labstack/echo/v5"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	UserV1Service *userV1.Service
)

func initService(app *Application) {
	UserV1Service = userV1.NewService(userV1.NewRepository(app.TimeLocation, app.Database))
}

func registeredFeatures() []dependency.Feature {
	return []dependency.Feature{
		userV1.Dependency(),
	}
}

func Start() error {
	var tag string = "internal.application.bootstrap.Start."

	cfg, err := config.New()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to initiate configuration")

		return err
	}

	timeLocation, err := time.LoadLocation(cfg.AppLocation)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to load location for time")

		return err
	}

	if err := dependency.Validate(registeredFeatures(), dependency.Availability{
		dependency.Database:      cfg.UseDatabase,
		dependency.Cache:         cfg.UseCache,
		dependency.MessageBroker: cfg.UseMessageBroker,
		dependency.ObjectStorage: cfg.UseObjectStorage,
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "03",
			"error": err.Error(),
		}).Error("failed to fulfill feature dependency")

		return err
	}

	var databaseConnection *gorm.DB

	if cfg.UseDatabase {
		databaseConnection, err = database.New(&database.Database{
			Connection:     cfg.DatabaseConnection,
			Host:           cfg.DatabaseHost,
			Port:           cfg.DatabasePort,
			Username:       cfg.DatabaseUsername,
			Password:       cfg.DatabasePassword,
			Name:           cfg.DatabaseName,
			SSLMode:        cfg.DatabaseSSLMode,
			ParseTime:      cfg.DatabaseParseTime,
			Charset:        cfg.DatabaseCharset,
			Timezone:       cfg.DatabaseTimezone,
			ConnectTimeout: cfg.DatabaseConnectTimeout,
		}, cfg.AppDebug)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "04",
				"error": err.Error(),
			}).Error("failed to connect database")

			return err
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
				"tag":   tag + "05",
				"error": err.Error(),
			}).Error("failed to connect cache")

			return err
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
				"tag":   tag + "06",
				"error": err.Error(),
			}).Error("failed to connect object storage")

			return err
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
				"tag":   tag + "07",
				"error": err.Error(),
			}).Error("failed to connect message broker")

			return err
		}
	}

	app := &Application{
		Config:        cfg,
		TimeLocation:  timeLocation,
		Database:      databaseConnection,
		Cache:         cacheConnection,
		ObjectStorage: objectStorageConnection,
		MessageBroker: messageBrokerConnection,
	}

	defer app.Close()

	initService(app)

	e := newServer(cfg)

	RegisterRoutes(e, app)

	return run(e, cfg)
}

func run(e *echo.Echo, cfg *config.Config) error {
	var tag string = "internal.application.bootstrap.run."

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	startConfig := echo.StartConfig{
		Address:         ":" + cfg.AppPort,
		GracefulTimeout: time.Duration(cfg.AppShutdownTimeout) * time.Second,
		OnShutdownError: func(err error) {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "01",
				"error": err.Error(),
			}).Error("failed to drain in flight requests within the shutdown timeout")
		},
	}

	if err := startConfig.Start(ctx, e); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to serve http")

		return err
	}

	logrus.WithFields(logrus.Fields{
		"tag": tag + "03",
	}).Info("server stopped, releasing connections")

	return nil
}
