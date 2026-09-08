package health

import (
	"net/http"

	"github.com/MrAndreID/goapi/v2/internal/entity"

	"github.com/labstack/echo/v5"
	"github.com/sirupsen/logrus"
)

type handler struct {
	Service InterfaceService
}

func NewHandler(e *echo.Echo, service InterfaceService) *handler {
	handler := &handler{
		Service: service,
	}

	e.GET("/health", handler.Check)

	return handler
}

func (h *handler) Check(c *echo.Context) error {
	if err := h.Service.Check(c.Request().Context()); err != nil {
		logrus.WithFields(logrus.Fields{
			"tag":   "internal.feature.health.handler.Check.01",
			"error": err.Error(),
		}).Error("failed to health check (from health service)")

		return c.JSON(entity.MainResponse{
			Code:  http.StatusServiceUnavailable,
			Error: []any{ErrServiceUnhealthy.Error()},
		}.JSON())
	}

	return c.JSON(entity.MainResponse{
		Code: http.StatusOK,
	}.JSON())
}
