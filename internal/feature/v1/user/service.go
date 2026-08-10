package user

import (
	"context"
	"errors"

	"github.com/MrAndreID/goapi/v2/internal/entity"

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
	Create(CreateData) (User, error)
	Read(context.Context, ReadData) (entity.PaginatorResponse, error)
	Update(UpdateData) error
	Delete(DeleteData) error
}

func (s *Service) Create(req CreateData) (User, error) {
	var (
		tag  string = "internal.feature.v1.user.service.Create."
		user User
	)

	for i := 0; i < len(req.Emails); i++ {
		for j := i + 1; j < len(req.Emails); j++ {
			if req.Emails[i] == req.Emails[j] {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "01",
					"error": "Duplicate Email",
				}).Error("duplicate email")

				return user, errors.New("DUPLICATE_EMAIL")
			}
		}
	}

	user, err := s.Repository.Create(req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to create user (from user repository)")

		return user, err
	}

	return user, nil
}

func (s *Service) Read(ctx context.Context, req ReadData) (entity.PaginatorResponse, error) {
	var (
		tag string = "internal.feature.v1.user.service.Read."
		err error
	)

	data, err := s.Repository.Read(ctx, req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to get user (from user repository)")

		return data, err
	}

	return data, nil
}

func (s *Service) Update(req UpdateData) error {
	var tag string = "internal.feature.v1.user.service.Update."

	if len(req.Emails) > 0 {
		for i := 0; i < len(req.Emails); i++ {
			for j := i + 1; j < len(req.Emails); j++ {
				if req.Emails[i] == req.Emails[j] {
					logrus.WithFields(logrus.Fields{
						"tag":   tag + "01",
						"error": "Duplicate Email",
					}).Error("duplicate email")

					return errors.New("DUPLICATE_EMAIL")
				}
			}
		}
	}

	err := s.Repository.Update(req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to update user (from user repository)")

		return err
	}

	return nil
}

func (s *Service) Delete(req DeleteData) error {
	var tag string = "internal.feature.v1.user.service.Delete."

	if err := s.Repository.Delete(req); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to delete user (from user repository)")

		return err
	}

	return nil
}
