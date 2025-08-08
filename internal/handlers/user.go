package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MrAndreID/goapi/internal/services"
	"github.com/MrAndreID/goapi/internal/types"

	"github.com/MrAndreID/gopackage"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type userHandler struct {
	UserService services.IUserService
}

func NewUserHandler(e *echo.Group, userService services.IUserService) *userHandler {
	handler := &userHandler{
		UserService: userService,
	}

	e.POST("/user", handler.Create)
	e.GET("/user", handler.Read)
	e.PATCH("/user/:id", handler.Update)
	e.DELETE("/user/:id", handler.Delete)

	return handler
}

func (h *userHandler) Create(c echo.Context) error {
	var (
		tag string = "internal.handlers.user.Create."
		req types.CreateUserRequest
	)

	if err := gopackage.EchoBindRequest(c, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.(*echo.HTTPError).Message,
		}).Error("invalid request data")

		return c.JSON(http.StatusBadRequest, types.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusBadRequest),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusBadRequest), " ", "_")),
			Data:        err.(*echo.HTTPError).Message,
		})
	}

	user, err := h.UserService.Create(types.CreateUserRequest{
		Name:   req.Name,
		Emails: req.Emails,
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to create user (from user service)")

		return c.JSON(http.StatusInternalServerError, types.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusInternalServerError),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusInternalServerError), " ", "_")),
		})
	}

	return c.JSON(http.StatusCreated, types.MainResponse{
		Code:        fmt.Sprintf("%04d", http.StatusCreated),
		Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusCreated), " ", "_")),
		Data:        user,
	})
}

func (h *userHandler) Read(c echo.Context) error {
	var (
		tag string = "internal.handlers.user.Read."
		req types.ReadUserRequest
	)

	if err := gopackage.EchoBindRequest(c, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.(*echo.HTTPError).Message,
		}).Error("invalid request data")

		return c.JSON(http.StatusBadRequest, types.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusBadRequest),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusBadRequest), " ", "_")),
			Data:        err.(*echo.HTTPError).Message,
		})
	}

	userData, err := h.UserService.Read(c.Request().Context(), req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to get user (from user service)")

		return c.JSON(http.StatusInternalServerError, types.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusInternalServerError),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusInternalServerError), " ", "_")),
		})
	}

	return c.JSON(http.StatusOK, types.MainResponse{
		Code:        fmt.Sprintf("%04d", http.StatusOK),
		Description: "SUCCESS",
		Data:        userData,
	})
}

func (h *userHandler) Update(c echo.Context) error {
	var (
		tag string = "internal.handlers.user.Update."
		req types.UpdateUserRequest
	)

	if err := gopackage.EchoBindRequest(c, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.(*echo.HTTPError).Message,
		}).Error("invalid request data")

		return c.JSON(http.StatusBadRequest, types.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusBadRequest),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusBadRequest), " ", "_")),
			Data:        err.(*echo.HTTPError).Message,
		})
	}

	err := h.UserService.Update(req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to update user (from user service)")

		return c.JSON(http.StatusInternalServerError, types.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusInternalServerError),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusInternalServerError), " ", "_")),
		})
	}

	return c.JSON(http.StatusOK, types.MainResponse{
		Code:        fmt.Sprintf("%04d", http.StatusOK),
		Description: "SUCCESS",
	})
}

func (h *userHandler) Delete(c echo.Context) error {
	var (
		tag string = "internal.handlers.user.Delete."
		req types.DeleteUserRequest
	)

	if err := gopackage.EchoBindRequest(c, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.(*echo.HTTPError).Message,
		}).Error("invalid request data")

		return c.JSON(http.StatusBadRequest, types.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusBadRequest),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusBadRequest), " ", "_")),
			Data:        err.(*echo.HTTPError).Message,
		})
	}

	if err := h.UserService.Delete(req.ID); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to delete user (from user service)")

		return c.JSON(http.StatusInternalServerError, types.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusInternalServerError),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusInternalServerError), " ", "_")),
		})
	}

	return c.JSON(http.StatusOK, types.MainResponse{
		Code:        fmt.Sprintf("%04d", http.StatusOK),
		Description: "SUCCESS",
	})
}
