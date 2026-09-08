package user

import (
	"net/http"

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

func respondError(err error) (int, entity.MainResponse) {
	status := StatusFor(err)

	response := entity.MainResponse{Code: status}

	if status < http.StatusInternalServerError {
		response.Error = []any{err.Error()}
	}

	return response.JSON()
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

		return c.JSON(entity.MainResponse{
			Code:  http.StatusBadRequest,
			Error: gopackage.ParseValidationErrors(err),
		}.JSON())
	}

	user, err := h.Service.Create(c.Request().Context(), req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to create user (from user service)")

		return c.JSON(respondError(err))
	}

	return c.JSON(entity.MainResponse{
		Code: http.StatusCreated,
		Data: user,
	}.JSON())
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

		return c.JSON(entity.MainResponse{
			Code:  http.StatusBadRequest,
			Error: gopackage.ParseValidationErrors(err),
		}.JSON())
	}

	userData, err := h.Service.Read(c.Request().Context(), req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to get user (from user service)")

		return c.JSON(respondError(err))
	}

	return c.JSON(entity.MainResponse{
		Code: http.StatusOK,
		Data: userData.Records,
		Meta: userData,
	}.JSON())
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

		return c.JSON(entity.MainResponse{
			Code:  http.StatusBadRequest,
			Error: gopackage.ParseValidationErrors(err),
		}.JSON())
	}

	err := h.Service.Update(c.Request().Context(), req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to update user (from user service)")

		return c.JSON(respondError(err))
	}

	return c.JSON(entity.MainResponse{
		Code: http.StatusOK,
	}.JSON())
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

		return c.JSON(entity.MainResponse{
			Code:  http.StatusBadRequest,
			Error: gopackage.ParseValidationErrors(err),
		}.JSON())
	}

	if err := h.Service.Delete(c.Request().Context(), req); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   tag + "02",
			"error": err.Error(),
		}).Error("failed to delete user (from user service)")

		return c.JSON(respondError(err))
	}

	return c.JSON(entity.MainResponse{
		Code: http.StatusOK,
	}.JSON())
}
