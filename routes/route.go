package routes

import (
	"todo-list-api/controllers"
	"todo-list-api/middlewares"

	"github.com/labstack/echo/v4"
)

func UserRoute(e *echo.Echo, controller controllers.UserController) {
	e.POST("/register", controller.Register)
	e.POST("/login", controller.Login)
	e.POST("/refresh-token", controller.RefershToken)
}

func TodoRoute(e *echo.Echo, controller controllers.TodoController) {
	e.POST("/todos", controller.Create, middlewares.Authenticate)
	e.PUT("/todos/:id", controller.Update, middlewares.Authenticate)
	e.DELETE("/todos/:id", controller.Delete, middlewares.Authenticate)
	e.GET("/todos", controller.FindWithPagination, middlewares.Authenticate)
}
