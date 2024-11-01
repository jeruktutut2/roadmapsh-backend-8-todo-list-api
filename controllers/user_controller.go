package controllers

import (
	"net/http"
	modelrequests "todo-list-api/models/requests"
	"todo-list-api/services"

	"github.com/labstack/echo/v4"
)

type UserController interface {
	Register(c echo.Context) error
	Login(c echo.Context) error
	RefershToken(c echo.Context) error
}

type UserControllerImplementation struct {
	UserService services.UserService
}

func NewUserController(userService services.UserService) UserController {
	return &UserControllerImplementation{
		UserService: userService,
	}
}

func (controller *UserControllerImplementation) Register(c echo.Context) error {
	var registerRequest modelrequests.RegisterRequest
	err := c.Bind(&registerRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}
	httpCode, accessToken, refreshToken, response := controller.UserService.Register(c.Request().Context(), registerRequest)

	if accessToken != "" {
		cookie := new(http.Cookie)
		cookie.Name = "Authorization"
		cookie.Value = accessToken
		c.SetCookie(cookie)
	}

	if refreshToken != "" {
		cookie := new(http.Cookie)
		cookie.Name = "refreshToken"
		cookie.Value = refreshToken
		c.SetCookie(cookie)
	}

	return c.JSON(httpCode, response)
}

func (controller *UserControllerImplementation) Login(c echo.Context) error {
	var loginRequest modelrequests.LoginRequest
	err := c.Bind(&loginRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}
	httpCode, accessToken, refreshToken, response := controller.UserService.Login(c.Request().Context(), loginRequest)

	cookie := new(http.Cookie)
	cookie.Name = "Authorization"
	cookie.Value = accessToken
	c.SetCookie(cookie)

	cookie = new(http.Cookie)
	cookie.Name = "refreshToken"
	cookie.Value = refreshToken
	c.SetCookie(cookie)

	return c.JSON(httpCode, response)
}

func (controller *UserControllerImplementation) RefershToken(c echo.Context) error {
	refreshTokenCookie, err := c.Cookie("refreshToken")
	if err != nil && err != http.ErrNoCookie {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	} else if err != nil && err == http.ErrNoCookie {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "cookie not found",
		})
	}

	httpCode, accessToken, response := controller.UserService.RefreshToken(c.Request().Context(), refreshTokenCookie.Value)
	cookie := new(http.Cookie)
	cookie.Name = "accessToken"
	cookie.Value = accessToken
	c.SetCookie(cookie)
	return c.JSON(httpCode, response)
}
