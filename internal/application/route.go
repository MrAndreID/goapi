package application

import (
	"github.com/MrAndreID/goapi/v2/internal/feature/v1/user"

	"github.com/MrAndreID/gomiddleware/v2"
	"github.com/labstack/echo/v5"
)

func RegisterRoutes(e *echo.Echo, app *Application) any {
	api := e.Group("/api", gomiddleware.EchoCheckApplicationKey(app.Config.AppKey))

	v1 := api.Group("/v1")
	user.NewHandler(v1, UserV1Service)

	return v1
}
