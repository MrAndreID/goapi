package version

import (
	"net/http"

	"github.com/MrAndreID/goapi/v2/internal/entity"

	"github.com/labstack/echo/v5"
)

type handler struct {
	Service InterfaceService
}

func NewHandler(e *echo.Echo, service InterfaceService) *handler {
	handler := &handler{
		Service: service,
	}

	e.GET("/version", handler.Read)

	return handler
}

func (h *handler) Read(c *echo.Context) error {
	return c.JSON(entity.MainResponse{
		Code: http.StatusOK,
		Data: h.Service.Read(c.Request().Context()),
	}.JSON())
}
