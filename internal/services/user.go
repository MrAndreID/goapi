package services

import (
	"context"
	"errors"
	"strconv"

	"github.com/MrAndreID/goapi/databases/models"
	"github.com/MrAndreID/goapi/internal/repositories"
	"github.com/MrAndreID/goapi/internal/types"

	"github.com/sirupsen/logrus"
)

type IUserService interface {
	Create(types.CreateUserRequest) (models.User, error)
	Read(context.Context, types.ReadUserRequest) (types.PaginatorResponse, error)
	Update(types.UpdateUserRequest) error
	Delete(string) error
}

type UserService struct {
	UserRepository repositories.IUserRepository
}

func NewUserService(userRepository repositories.IUserRepository) *UserService {
	return &UserService{
		UserRepository: userRepository,
	}
}

func (s *UserService) Create(req types.CreateUserRequest) (models.User, error) {
	var (
		tag  string = "internal.services.user.Create."
		user models.User
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

	user, err := s.UserRepository.Create(repositories.CreateUserData{
		Name:   req.Name,
		Emails: req.Emails,
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to create user (from user repository)")

		return user, err
	}

	return user, nil
}

func (s *UserService) Read(ctx context.Context, req types.ReadUserRequest) (types.PaginatorResponse, error) {
	var (
		tag                   string = "internal.services.user.Read."
		res                   types.PaginatorResponse
		err                   error
		page, limit           int
		disableCalculateTotal bool
	)

	if req.Page != "" {
		page, err = strconv.Atoi(req.Page)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "01",
				"error": err.Error(),
			}).Error("failed to convert from string to int for page from request")

			return res, err
		}
	}

	if req.Limit != "" {
		limit, err = strconv.Atoi(req.Limit)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "02",
				"error": err.Error(),
			}).Error("failed to convert from string to int for limit from request")

			return res, err
		}
	}

	if req.DisableCalculateTotal != "" {
		disableCalculateTotal, err = strconv.ParseBool(req.DisableCalculateTotal)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "03",
				"error": err.Error(),
			}).Error("failed to convert from string to bool for disable calculate total from request")

			return res, err
		}
	}

	data, err := s.UserRepository.Read(ctx, repositories.ReadUserData{
		Page:                  page,
		Limit:                 limit,
		OrderBy:               req.OrderBy,
		SortBy:                req.SortBy,
		Search:                req.Search,
		DisableCalculateTotal: disableCalculateTotal,
		ID:                    req.ID,
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "04",
			"error": err.Error(),
		}).Error("failed to get user (from user repository)")

		return data, err
	}

	return data, nil
}

func (s *UserService) Update(req types.UpdateUserRequest) error {
	var tag string = "internal.services.user.Update."

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

	err := s.UserRepository.Update(repositories.UpdateUserData{
		ID:     req.ID,
		Name:   req.Name,
		Emails: req.Emails,
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to update user (from user repository)")

		return err
	}

	return nil
}

func (s *UserService) Delete(id string) error {
	var tag string = "internal.services.user.Delete."

	if err := s.UserRepository.Delete(id); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to delete user (from user repository)")

		return err
	}

	return nil
}
