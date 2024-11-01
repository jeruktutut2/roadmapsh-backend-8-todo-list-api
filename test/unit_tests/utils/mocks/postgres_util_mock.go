package mockutils

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/mock"
)

type PostgresUtilMock struct {
	Mock mock.Mock
}

func (util *PostgresUtilMock) GetPool() *pgxpool.Pool {
	arguments := util.Mock.Called()
	return arguments.Get(0).(*pgxpool.Pool)
}

func (util *PostgresUtilMock) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	arguments := util.Mock.Called(ctx, options)
	return arguments.Get(0).(pgx.Tx), arguments.Error(1)
}

func (util *PostgresUtilMock) Close() {
	fmt.Println("Close")
}

func (util *PostgresUtilMock) CommitOrRollback(tx pgx.Tx, ctx context.Context, err error) error {
	arguments := util.Mock.Called(tx, ctx, err)
	return arguments.Error(0)
}
