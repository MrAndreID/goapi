package health

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Service struct {
	Repository InterfaceRepository
}

func NewService(repository InterfaceRepository) *Service {
	return &Service{
		Repository: repository,
	}
}

type InterfaceService interface {
	Check(context.Context) error
}

func (s *Service) Check(ctx context.Context) error {
	var tag string = "internal.feature.health.service.Check."

	databaseStatus, err := s.Repository.CheckDatabase(ctx)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to check database (from health repository)")

		return err
	}

	if !databaseStatus {
		logrus.WithFields(logrus.Fields{
			"tag": tag + "02",
		}).Info("database status failed")

		return ErrDatabaseUnavailable
	}

	cacheStatus, err := s.Repository.CheckCache(ctx)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "03",
			"error": err.Error(),
		}).Error("failed to check cache (from health repository)")

		return err
	}

	if !cacheStatus {
		logrus.WithFields(logrus.Fields{
			"tag": tag + "04",
		}).Info("cache status failed")

		return ErrCacheUnavailable
	}

	messageBrokerStatus, err := s.Repository.CheckMessageBroker(ctx)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "05",
			"error": err.Error(),
		}).Error("failed to check message broker (from health repository)")

		return err
	}

	if !messageBrokerStatus {
		logrus.WithFields(logrus.Fields{
			"tag": tag + "06",
		}).Info("message broker status failed")

		return ErrMessageBrokerUnavailable
	}

	objectStorageStatus, err := s.Repository.CheckObjectStorage(ctx)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "07",
			"error": err.Error(),
		}).Error("failed to check object storage (from health repository)")

		return err
	}

	if !objectStorageStatus {
		logrus.WithFields(logrus.Fields{
			"tag": tag + "08",
		}).Info("object storage status failed")

		return ErrObjectStorageUnavailable
	}

	return nil
}
