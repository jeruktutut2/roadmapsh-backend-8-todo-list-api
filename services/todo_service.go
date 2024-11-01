package services

import (
	"context"
	"net/http"
	"todo-list-api/helpers"
	modelentities "todo-list-api/models/entities"
	modelrequests "todo-list-api/models/requests"
	modelresponses "todo-list-api/models/responses"
	"todo-list-api/repositories"
	"todo-list-api/utils"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TodoService interface {
	Create(ctx context.Context, createTodoRequest modelrequests.CreateTodoRequest) (httpCode int, response interface{})
	Update(ctx context.Context, id int, updateTodoRequest modelrequests.UpdateTodoRequest) (httpCode int, response interface{})
	Delete(ctx context.Context, id int) (httpCode int, response interface{})
	FindWithPagination(ctx context.Context, page int, limit int) (httpCode int, response interface{})
}

type TodoServiceImplementation struct {
	PostgresUtil   utils.PostgresUtil
	Validate       *validator.Validate
	TodoRepository repositories.TodoRepository
}

func NewTodoService(postgresUtil utils.PostgresUtil, validate *validator.Validate, todoRepository repositories.TodoRepository) TodoService {
	return &TodoServiceImplementation{
		PostgresUtil:   postgresUtil,
		Validate:       validate,
		TodoRepository: todoRepository,
	}
}

func (service *TodoServiceImplementation) Create(ctx context.Context, createTodoRequest modelrequests.CreateTodoRequest) (httpCode int, response interface{}) {
	err := service.Validate.Struct(createTodoRequest)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	userId, ok := ctx.Value("userId").(int)
	if !ok {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse("cannot find user id")
		return
	}

	tx, err := service.PostgresUtil.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	defer func() {
		errCommitOrRollback := service.PostgresUtil.CommitOrRollback(tx, ctx, err)
		if errCommitOrRollback != nil {
			httpCode = http.StatusInternalServerError
			response = helpers.ToResponse(errCommitOrRollback.Error())
		}
	}()

	var todo modelentities.Todo
	todo.UserId = pgtype.Int4{Valid: true, Int32: int32(userId)}
	todo.Title = pgtype.Text{Valid: true, String: createTodoRequest.Title}
	todo.Description = pgtype.Text{Valid: true, String: createTodoRequest.Description}
	lastInsertedId, err := service.TodoRepository.Create(tx, ctx, todo)
	if err != nil && err != pgx.ErrNoRows {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	} else if err != nil && err == pgx.ErrNoRows {
		httpCode = http.StatusBadRequest
		response = helpers.ToResponse(err.Error())
		return
	}
	todo.Id = pgtype.Int4{Valid: true, Int32: int32(lastInsertedId)}

	var createTodoResponse modelresponses.CreateTodoResponse
	createTodoResponse.Id = int(todo.Id.Int32)
	createTodoResponse.Title = todo.Title.String
	createTodoResponse.Description = todo.Description.String

	httpCode = http.StatusCreated
	response = createTodoResponse
	return
}

func (service *TodoServiceImplementation) Update(ctx context.Context, id int, updateTodoRequest modelrequests.UpdateTodoRequest) (httpCode int, response interface{}) {
	err := service.Validate.Struct(updateTodoRequest)
	if err != nil {
		httpCode = http.StatusBadRequest
		response = helpers.ToResponse(err.Error())
		return
	}

	userId, ok := ctx.Value("userId").(int)
	if !ok {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse("cannot find user id")
		return
	}

	tx, err := service.PostgresUtil.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		httpCode = http.StatusInsufficientStorage
		response = helpers.ToResponse(err.Error())
		return
	}

	defer func() {
		errCommitOrRollback := service.PostgresUtil.CommitOrRollback(tx, ctx, err)
		if errCommitOrRollback != nil {
			httpCode = http.StatusInternalServerError
			response = helpers.ToResponse(errCommitOrRollback.Error())
		}
	}()

	todo, err := service.TodoRepository.FindByIdAndUserId(tx, ctx, id, userId)
	if err != nil && err != pgx.ErrNoRows {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	} else if err != nil && err == pgx.ErrNoRows {
		httpCode = http.StatusForbidden
		response = helpers.ToResponse("forbidden")
		return
	}

	todo.Title = pgtype.Text{Valid: true, String: updateTodoRequest.Title}
	todo.Description = pgtype.Text{Valid: true, String: updateTodoRequest.Description}
	rowsAffected, err := service.TodoRepository.Update(tx, ctx, todo)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	if rowsAffected != 1 {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse("rows affected not one")
		return
	}

	var updateTodoResponse modelresponses.UpdateTodoResponse
	updateTodoResponse.Id = int(todo.Id.Int32)
	updateTodoResponse.Title = todo.Title.String
	updateTodoResponse.Description = todo.Description.String

	httpCode = http.StatusOK
	response = updateTodoResponse
	return
}

func (service *TodoServiceImplementation) Delete(ctx context.Context, id int) (httpCode int, response interface{}) {
	tx, err := service.PostgresUtil.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	defer func() {
		errCommitOrRollback := service.PostgresUtil.CommitOrRollback(tx, ctx, err)
		if errCommitOrRollback != nil {
			httpCode = http.StatusInternalServerError
			response = helpers.ToResponse(errCommitOrRollback.Error())
		}
	}()

	userId, ok := ctx.Value("userId").(int)
	if !ok {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse("cannot find user id")
		return
	}

	_, err = service.TodoRepository.FindByIdAndUserId(tx, ctx, id, userId)
	if err != nil && err != pgx.ErrNoRows {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	} else if err != nil && err == pgx.ErrNoRows {
		httpCode = http.StatusForbidden
		response = helpers.ToResponse("forbidden")
		return
	}

	rowsAffected, err := service.TodoRepository.Delete(tx, ctx, id)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	if rowsAffected != 1 {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse("rows affected not one")
		return
	}

	httpCode = http.StatusNoContent
	response = helpers.ToResponse("successfully deleted")
	return
}

func (service *TodoServiceImplementation) FindWithPagination(ctx context.Context, page int, limit int) (httpCode int, response interface{}) {
	userId, ok := ctx.Value("userId").(int)
	if !ok {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse("cannot find user id")
		return
	}
	offset := (page - 1) * limit
	todos, err := service.TodoRepository.FindByPagination(service.PostgresUtil.GetPool(), ctx, userId, offset, limit)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	if len(todos) < 1 {
		httpCode = http.StatusNotFound
		response = helpers.ToResponse("cannot find todos")
		return
	}

	numberOfTodos, err := service.TodoRepository.Count(service.PostgresUtil.GetPool(), ctx)
	if err != nil && err != pgx.ErrNoRows {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	} else if err != nil && err == pgx.ErrNoRows {
		httpCode = http.StatusBadRequest
		response = helpers.ToResponse(err.Error())
		return
	}

	var todoResponses []modelresponses.TodoResponse
	for _, todo := range todos {
		var todoResponse modelresponses.TodoResponse
		todoResponse.Id = int(todo.Id.Int32)
		todoResponse.Title = todo.Title.String
		todoResponse.Description = todo.Description.String
		todoResponses = append(todoResponses, todoResponse)
	}

	var getTodoResponse modelresponses.GetTodoResponse
	getTodoResponse.Data = todoResponses
	getTodoResponse.Page = page
	getTodoResponse.Limit = limit
	getTodoResponse.Total = numberOfTodos

	httpCode = http.StatusOK
	response = getTodoResponse
	return
}
