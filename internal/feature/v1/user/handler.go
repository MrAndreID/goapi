package user

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MrAndreID/goapi/v2/internal/entity"

	"github.com/MrAndreID/gopackage/v2"
	"github.com/labstack/echo/v5"
	"github.com/sirupsen/logrus"
)

type handler struct {
	Service InterfaceService
}

func NewHandler(e *echo.Group, service InterfaceService) *handler {
	handler := &handler{
		Service: service,
	}

	e.POST("/user", handler.Create)
	e.GET("/user", handler.Read)
	e.PATCH("/user/:id", handler.Update)
	e.DELETE("/user/:id", handler.Delete)

	return handler
}

func (h *handler) Create(c *echo.Context) error {
	var (
		tag string = "internal.feature.v1.user.handler.Create."
		req CreateData
	)

	if err := gopackage.EchoBindValidateRequest(c, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.(*echo.HTTPError).Message,
		}).Error("invalid request data")

		return c.JSON(http.StatusBadRequest, entity.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusBadRequest),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusBadRequest), " ", "_")),
			Data:        err.(*echo.HTTPError).Message,
		})
	}

	user, err := h.Service.Create(c.Request().Context(), req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to create user (from user service)")

		return c.JSON(http.StatusInternalServerError, entity.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusInternalServerError),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusInternalServerError), " ", "_")),
		})
	}

	return c.JSON(http.StatusCreated, entity.MainResponse{
		Code:        fmt.Sprintf("%04d", http.StatusCreated),
		Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusCreated), " ", "_")),
		Data:        user,
	})
}

func (h *handler) Read(c *echo.Context) error {
	var (
		tag string = "internal.feature.v1.user.handler.Read."
		req ReadData
	)

	if err := gopackage.EchoBindValidateRequest(c, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.(*echo.HTTPError).Message,
		}).Error("invalid request data")

		return c.JSON(http.StatusBadRequest, entity.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusBadRequest),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusBadRequest), " ", "_")),
			Data:        err.(*echo.HTTPError).Message,
		})
	}

	userData, err := h.Service.Read(c.Request().Context(), req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to get user (from user service)")

		return c.JSON(http.StatusInternalServerError, entity.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusInternalServerError),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusInternalServerError), " ", "_")),
		})
	}

	return c.JSON(http.StatusOK, entity.MainResponse{
		Code:        fmt.Sprintf("%04d", http.StatusOK),
		Description: "SUCCESS",
		Data:        userData,
	})
}

func (h *handler) Update(c *echo.Context) error {
	var (
		tag string = "internal.feature.v1.user.handler.Update."
		req UpdateData
	)

	if err := gopackage.EchoBindValidateRequest(c, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.(*echo.HTTPError).Message,
		}).Error("invalid request data")

		return c.JSON(http.StatusBadRequest, entity.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusBadRequest),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusBadRequest), " ", "_")),
			Data:        err.(*echo.HTTPError).Message,
		})
	}

	err := h.Service.Update(c.Request().Context(), req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to update user (from user service)")

		return c.JSON(http.StatusInternalServerError, entity.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusInternalServerError),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusInternalServerError), " ", "_")),
		})
	}

	return c.JSON(http.StatusOK, entity.MainResponse{
		Code:        fmt.Sprintf("%04d", http.StatusOK),
		Description: "SUCCESS",
	})
}

func (h *handler) Delete(c *echo.Context) error {
	var (
		tag string = "internal.feature.v1.user.handler.Delete."
		req DeleteData
	)

	if err := gopackage.EchoBindValidateRequest(c, &req); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "01",
			"error": err.(*echo.HTTPError).Message,
		}).Error("invalid request data")

		return c.JSON(http.StatusBadRequest, entity.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusBadRequest),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusBadRequest), " ", "_")),
			Data:        err.(*echo.HTTPError).Message,
		})
	}

	if err := h.Service.Delete(c.Request().Context(), req); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to delete user (from user service)")

		return c.JSON(http.StatusInternalServerError, entity.MainResponse{
			Code:        fmt.Sprintf("%04d", http.StatusInternalServerError),
			Description: strings.ToUpper(strings.ReplaceAll(http.StatusText(http.StatusInternalServerError), " ", "_")),
		})
	}

	return c.JSON(http.StatusOK, entity.MainResponse{
		Code:        fmt.Sprintf("%04d", http.StatusOK),
		Description: "SUCCESS",
	})
}
