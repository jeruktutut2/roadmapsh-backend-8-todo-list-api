package controllers

import (
	"net/http"
	"strconv"
	modelrequests "todo-list-api/models/requests"
	"todo-list-api/services"

	"github.com/labstack/echo/v4"
)

type TodoController interface {
	Create(c echo.Context) error
	Update(c echo.Context) error
	Delete(c echo.Context) error
	FindWithPagination(c echo.Context) error
}

type TodoControllerImplementation struct {
	TodoService services.TodoService
}

func NewTodoController(todoService services.TodoService) TodoController {
	return &TodoControllerImplementation{
		TodoService: todoService,
	}
}

func (controller *TodoControllerImplementation) Create(c echo.Context) error {
	var createTodoRequest modelrequests.CreateTodoRequest
	err := c.Bind(&createTodoRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}
	httpCode, response := controller.TodoService.Create(c.Request().Context(), createTodoRequest)
	return c.JSON(httpCode, response)
}

func (controller *TodoControllerImplementation) Update(c echo.Context) error {
	var updateTodoRequest modelrequests.UpdateTodoRequest
	err := c.Bind(&updateTodoRequest)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}
	httpCode, response := controller.TodoService.Update(c.Request().Context(), id, updateTodoRequest)
	return c.JSON(httpCode, response)
}

func (controller *TodoControllerImplementation) Delete(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}
	httpCode, response := controller.TodoService.Delete(c.Request().Context(), id)
	return c.JSON(httpCode, response)
}

func (controller *TodoControllerImplementation) FindWithPagination(c echo.Context) error {
	pageQueryParam := c.QueryParam("page")
	page, err := strconv.Atoi(pageQueryParam)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}
	limitQueryParam := c.QueryParam("limit")
	limit, err := strconv.Atoi(limitQueryParam)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	httpCode, respose := controller.TodoService.FindWithPagination(c.Request().Context(), page, limit)
	return c.JSON(httpCode, respose)
}
