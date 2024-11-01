package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
	"todo-list-api/controllers"
	"todo-list-api/helpers"
	"todo-list-api/middlewares"
	"todo-list-api/repositories"
	"todo-list-api/routes"
	"todo-list-api/services"
	"todo-list-api/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func main() {
	postgresUtil := utils.NewPostgresConnection()

	numberOfLimitEnv := os.Getenv("NUMBER_OF_LIMIT")
	numberOfLimit, err := strconv.Atoi(numberOfLimitEnv)
	if err != nil {
		panic(err.Error())
	}
	middlewares.NumberOfLimit = numberOfLimit

	e := echo.New()
	e.Use(middlewares.SetRateLimiter)
	validate := validator.New()
	bcryptHelper := helpers.NewBcryptHelper()
	jwtHelper := helpers.NewJwtHelper()

	userRepository := repositories.NewUserRepository()
	userService := services.NewUserService(postgresUtil, validate, userRepository, bcryptHelper, jwtHelper)
	userController := controllers.NewUserController(userService)
	routes.UserRoute(e, userController)

	todoRepository := repositories.NewTodoRepository()
	todoService := services.NewTodoService(postgresUtil, validate, todoRepository)
	todoController := controllers.NewTodoController(todoService)
	routes.TodoRoute(e, todoController)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		if err := e.Start(os.Getenv("ECHO_HOST")); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

}
