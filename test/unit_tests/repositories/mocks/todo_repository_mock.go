package mockrepositories

import (
	"context"

	modelentities "todo-list-api/models/entities"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/mock"
)

type TodoRepositoryMock struct {
	Mock mock.Mock
}

func (repository *TodoRepositoryMock) Create(tx pgx.Tx, ctx context.Context, todo modelentities.Todo) (lastInsertId int, err error) {
	arguments := repository.Mock.Called(tx, ctx, todo)
	return arguments.Get(0).(int), arguments.Error(1)
}

func (repository *TodoRepositoryMock) FindByIdAndUserId(tx pgx.Tx, ctx context.Context, id int, userId int) (todo modelentities.Todo, err error) {
	arguments := repository.Mock.Called(tx, ctx, id, userId)
	return arguments.Get(0).(modelentities.Todo), arguments.Error(1)
}

func (repository *TodoRepositoryMock) Update(tx pgx.Tx, ctx context.Context, todo modelentities.Todo) (rowsAffected int64, err error) {
	arguments := repository.Mock.Called(tx, ctx, todo)
	return arguments.Get(0).(int64), arguments.Error(1)
}

func (repository *TodoRepositoryMock) Delete(tx pgx.Tx, ctx context.Context, id int) (rowsAffected int64, err error) {
	arguments := repository.Mock.Called(tx, ctx, id)
	return arguments.Get(0).(int64), arguments.Error(1)
}

func (repository *TodoRepositoryMock) FindByPagination(pool *pgxpool.Pool, ctx context.Context, userId int, offset int, limit int) (todos []modelentities.Todo, err error) {
	arguments := repository.Mock.Called(pool, ctx, userId, offset, limit)
	return arguments.Get(0).([]modelentities.Todo), arguments.Error(1)
}

func (repository *TodoRepositoryMock) Count(pool *pgxpool.Pool, ctx context.Context) (numberOfTodos int, err error) {
	arguments := repository.Mock.Called(pool, ctx)
	return arguments.Get(0).(int), arguments.Error(1)
}
