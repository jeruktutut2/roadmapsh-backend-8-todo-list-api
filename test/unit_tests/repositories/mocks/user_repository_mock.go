package mockrepositories

import (
	"context"

	modelentities "todo-list-api/models/entities"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/mock"
)

type UserRepositoryMock struct {
	Mock mock.Mock
}

func (repository *UserRepositoryMock) Create(tx pgx.Tx, ctx context.Context, user modelentities.User) (lastInsertedId int, err error) {
	arguments := repository.Mock.Called(tx, ctx, user)
	return arguments.Get(0).(int), arguments.Error(1)
}

func (repository *UserRepositoryMock) UpdateRefreshToken(tx pgx.Tx, ctx context.Context, refreshToken string, id int) (rowsAffected int64, err error) {
	arguments := repository.Mock.Called(tx, ctx, refreshToken, id)
	return arguments.Get(0).(int64), arguments.Error(1)
}

func (repository *UserRepositoryMock) FindByEmail(tx pgx.Tx, ctx context.Context, email string) (user modelentities.User, err error) {
	arguments := repository.Mock.Called(tx, ctx, email)
	return arguments.Get(0).(modelentities.User), arguments.Error(1)
}

func (repository *UserRepositoryMock) FindByRefreshToken(pool *pgxpool.Pool, ctx context.Context, refreshToken string) (user modelentities.User, err error) {
	arguments := repository.Mock.Called(pool, ctx, refreshToken)
	return arguments.Get(0).(modelentities.User), arguments.Error(1)
}
