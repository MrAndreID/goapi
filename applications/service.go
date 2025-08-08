package applications

import (
	"github.com/MrAndreID/goapi/internal/repositories"
	"github.com/MrAndreID/goapi/internal/services"
)

var (
	UserService *services.UserService
)

func initService(app *Application) {
	UserService = services.NewUserService(repositories.NewUserRepository(app.TimeLocation, app.Database))
}
