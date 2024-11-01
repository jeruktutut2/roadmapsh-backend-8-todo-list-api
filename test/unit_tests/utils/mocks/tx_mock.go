package mockutils

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
)

type PgxTxMock struct {
	Mock mock.Mock
}

func (pgxTx *PgxTxMock) Begin(ctx context.Context) (pgx.Tx, error) {
	arguments := pgxTx.Mock.Called(ctx)
	return arguments.Get(0).(pgx.Tx), arguments.Error(1)
}

func (pgxTx *PgxTxMock) Commit(ctx context.Context) error {
	arguments := pgxTx.Mock.Called(ctx)
	return arguments.Error(0)
}

func (pgxTx *PgxTxMock) Rollback(ctx context.Context) error {
	arguments := pgxTx.Mock.Called(ctx)
	return arguments.Error(0)
}

func (pgxTx *PgxTxMock) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	arguments := pgxTx.Mock.Called(ctx, tableName, columnNames, rowSrc)
	return arguments.Get(0).(int64), arguments.Error(1)
}

func (pgxTx *PgxTxMock) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	arguments := pgxTx.Mock.Called(ctx, b)
	return arguments.Get(0).(pgx.BatchResults)
}

func (pgxTx *PgxTxMock) LargeObjects() pgx.LargeObjects {
	arguments := pgxTx.Mock.Called()
	return arguments.Get(0).(pgx.LargeObjects)
}

func (pgxTx *PgxTxMock) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	arguments := pgxTx.Mock.Called(ctx, name, sql)
	return arguments.Get(0).(*pgconn.StatementDescription), arguments.Error(1)
}

func (pgxTx *PgxTxMock) Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error) {
	args := pgxTx.Mock.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (pgxTx *PgxTxMock) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	arguments := pgxTx.Mock.Called(ctx, sql, args)
	return arguments.Get(0).(pgx.Rows), arguments.Error(1)
}

func (pgxTx *PgxTxMock) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	arguments := pgxTx.Mock.Called(ctx, sql, args)
	return arguments.Get(0).(pgx.Row)
}

func (pgxTx *PgxTxMock) Conn() *pgx.Conn {
	arguments := pgxTx.Mock.Called()
	return arguments.Get(0).(*pgx.Conn)
}
