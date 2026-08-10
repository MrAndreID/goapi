package user

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/MrAndreID/goapi/v2/internal/entity"

	"github.com/MrAndreID/gopackage/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository struct {
	TimeLocation *time.Location
	Database     *gorm.DB
}

func NewRepository(timeLocation *time.Location, db *gorm.DB) *Repository {
	return &Repository{
		TimeLocation: timeLocation,
		Database:     db,
	}
}

type InterfaceRepository interface {
	Create(CreateData) (User, error)
	Read(context.Context, ReadData) (entity.PaginatorResponse, error)
	Update(UpdateData) error
	Delete(DeleteData) error
}

func (r *Repository) Create(req CreateData) (User, error) {
	var (
		tag  string = "internal.feature.v1.user.repository.Create."
		user User
	)

	tx := r.Database.Begin()

	userUUID, err := uuid.NewRandom()

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.Error(),
		}).Error("failed to generate uuid")

		tx.Rollback()

		return user, err
	}

	user.ID = userUUID.String()
	user.CreatedAt = time.Now().In(r.TimeLocation)
	user.UpdatedAt = time.Now().In(r.TimeLocation)
	user.Name = req.Name

	createUser := tx.Save(&user)

	if createUser.Error != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": createUser.Error.Error(),
		}).Error("failed to create user")

		tx.Rollback()

		return user, createUser.Error
	}

	if createUser.RowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "03",
			"error": "Failed to Create User",
		}).Error("failed to create user")

		tx.Rollback()

		return user, errors.New("FAILED_TO_CREATE_USER")
	}

	for _, v := range req.Emails {
		var email Email

		emailUUID, err := uuid.NewRandom()

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "04",
				"error": err.Error(),
			}).Error("failed to generate uuid")

			tx.Rollback()

			return user, err
		}

		email.ID = emailUUID.String()
		email.CreatedAt = time.Now().In(r.TimeLocation)
		email.UpdatedAt = time.Now().In(r.TimeLocation)
		email.UserID = user.ID
		email.Email = v

		createEmail := tx.Save(&email)

		if createEmail.Error != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "05",
				"error": createEmail.Error.Error(),
			}).Error("failed to create email")

			tx.Rollback()

			return user, createEmail.Error
		}

		if createEmail.RowsAffected == 0 {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "06",
				"error": "Failed to Create Email",
			}).Error("failed to create email")

			tx.Rollback()

			return user, errors.New("FAILED_TO_CREATE_EMAIL")
		}

		user.Emails = append(user.Emails, email)
	}

	tx.Commit()

	return user, nil
}

func (r *Repository) Read(ctx context.Context, req ReadData) (entity.PaginatorResponse, error) {
	var (
		tag     string = "internal.feature.v1.user.repository.Read."
		users   []User
		orderBy map[string]string = map[string]string{
			"id":        "id",
			"name":      "name",
			"createdAt": "created_at",
			"updatedAt": "updated_at",
		}
		sortBy map[string]string = map[string]string{
			"asc":  "asc",
			"desc": "desc",
		}
		search                []string = []string{"name"}
		page, limit           int
		disableCalculateTotal bool
		err                   error
		total                 int64
		res                   entity.PaginatorResponse
	)

	if req.Page != "" {
		page, err = strconv.Atoi(req.Page)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "01",
				"error": err.Error(),
			}).Error("failed to convert from string to int for page")

			return res, err
		}
	}

	if req.Limit != "" {
		limit, err = strconv.Atoi(req.Limit)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "02",
				"error": err.Error(),
			}).Error("failed to convert from string to int for limit")

			return res, err
		}
	}

	if req.DisableCalculateTotal != "" {
		disableCalculateTotal, err = strconv.ParseBool(req.DisableCalculateTotal)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "03",
				"error": err.Error(),
			}).Error("failed to convert from string to bool for disable calculate total")

			return res, err
		}
	}

	countTotal := r.Database.Model(&User{}).Preload("Emails")

	queryBuilder := r.Database.Model(&User{}).Preload("Emails")

	if req.ID != "" {
		countTotal.Where("id = ?", req.ID)

		queryBuilder.Where("id = ?", req.ID)
	}

	gopackage.DataTable(
		ctx,
		queryBuilder,
		search,
		orderBy[req.OrderBy],
		sortBy[req.SortBy],
		orderBy["name"],
		sortBy["asc"],
		page,
		&limit,
		req.Search,
		false,
	)

	queryBuilder.Find(&users)

	res.Records = users

	if !disableCalculateTotal {
		countTotal.Count(&total)

		res.Total = total
	}

	if len(users) >= limit {
		res.NextPage = true
	}

	return res, nil
}

func (r *Repository) Update(req UpdateData) error {
	var (
		tag  string = "internal.feature.v1.user.repository.Update."
		user User
	)

	tx := r.Database.Begin()

	readUser := tx.First(&user, "id = ?", req.ID)

	if readUser.RowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": "Failed to Read User Data",
		}).Error("failed to read user data")

		tx.Rollback()

		return errors.New("FAILED_TO_READ_USER_DATA")
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if len(req.Emails) > 0 {
		deleteEmail := tx.Where("user_id = ?", user.ID).Delete(&Email{})

		if deleteEmail.Error != nil {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "02",
				"error": deleteEmail.Error.Error(),
			}).Error("failed to delete email data")

			tx.Rollback()

			return deleteEmail.Error
		}

		if deleteEmail.RowsAffected == 0 {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "03",
				"error": "Failed to Delete Email Data",
			}).Error("failed to delete email data")

			tx.Rollback()

			return errors.New("FAILED_TO_DELETE_EMAIL_DATA")
		}

		for _, v := range req.Emails {
			var email Email

			emailUUID, err := uuid.NewRandom()

			if err != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "04",
					"error": err.Error(),
				}).Error("failed to generate uuid")

				tx.Rollback()

				return err
			}

			email.ID = emailUUID.String()
			email.CreatedAt = time.Now().In(r.TimeLocation)
			email.UpdatedAt = time.Now().In(r.TimeLocation)
			email.UserID = user.ID
			email.Email = v

			createEmail := tx.Save(&email)

			if createEmail.Error != nil {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "05",
					"error": createEmail.Error.Error(),
				}).Error("failed to create email")

				tx.Rollback()

				return createEmail.Error
			}

			if createEmail.RowsAffected == 0 {
				logrus.WithFields(logrus.Fields{
					"tag":   tag + "06",
					"error": "Failed to Create Email",
				}).Error("failed to create email")

				tx.Rollback()

				return errors.New("FAILED_TO_CREATE_EMAIL")
			}

			user.Emails = append(user.Emails, email)
		}
	} else {
		var emails []Email

		readEmail := tx.Find(&emails, "user_id = ?", user.ID)

		if readEmail.RowsAffected == 0 {
			logrus.WithFields(logrus.Fields{
				"tag":   tag + "07",
				"error": "Failed to Read Email Data",
			}).Error("failed to read email data")

			tx.Rollback()

			return errors.New("FAILED_TO_READ_EMAIL_DATA")
		}

		user.Emails = emails
	}

	user.UpdatedAt = time.Now().In(r.TimeLocation)

	updateUser := tx.Save(&user)

	if updateUser.Error != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "08",
			"error": updateUser.Error.Error(),
		}).Error("failed to update user data")

		tx.Rollback()

		return updateUser.Error
	}

	if updateUser.RowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "09",
			"error": "Failed to Update User Data",
		}).Error("failed to update user data")

		tx.Rollback()

		return errors.New("FAILED_TO_UPDATE_USER_DATA")
	}

	tx.Commit()

	return nil
}

func (r *Repository) Delete(req DeleteData) error {
	var (
		tag  string = "internal.feature.v1.user.repository.Delete."
		user User
	)

	tx := r.Database.Begin()

	readUser := tx.First(&user, "id = ?", req.ID)

	if readUser.RowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": "Failed To Read User Data",
		}).Error("failed to read user data")

		tx.Rollback()

		return errors.New("FAILED_TO_READ_USER_DATA")
	}

	deleteUser := tx.Delete(&user, "id = ?", req.ID)

	if deleteUser.Error != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": deleteUser.Error.Error(),
		}).Error("failed to delete user data")

		tx.Rollback()

		return deleteUser.Error
	}

	if deleteUser.RowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "03",
			"error": "Failed To Delete User Data",
		}).Error("failed to delete user data")

		tx.Rollback()

		return errors.New("FAILED_TO_DELETE_USER_DATA")
	}

	deleteEmail := tx.Where("user_id = ?", req.ID).Delete(&Email{})

	if deleteEmail.Error != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "04",
			"error": deleteEmail.Error.Error(),
		}).Error("failed to delete email data")

		tx.Rollback()

		return deleteEmail.Error
	}

	if deleteEmail.RowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "05",
			"error": "Failed To Delete Email Data",
		}).Error("failed to delete email data")

		tx.Rollback()

		return errors.New("FAILED_TO_DELETE_EMAIL_DATA")
	}

	tx.Commit()

	return nil
}
