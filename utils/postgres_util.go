package utils

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUtil interface {
	GetPool() *pgxpool.Pool
	BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
	Close()
	CommitOrRollback(tx pgx.Tx, ctx context.Context, err error) error
}

type PostgresUtilImplementation struct {
	pool *pgxpool.Pool
}

func NewPostgresConnection() PostgresUtil {
	println(time.Now().String(), "postgres: connecting to", os.Getenv("POSTGRES_HOST"))
	ctx := context.Background()
	connectionString := "postgres://" + os.Getenv("POSTGRES_USERNAME") + ":" + os.Getenv("POSTGRES_PASSWORD") + "@" + os.Getenv("POSTGRES_HOST") + "/" + os.Getenv("POSTGRES_DATABASE")
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		log.Fatalln("error when parse config: " + err.Error())
	}

	maxConnection, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_CONNECTION"))
	if err != nil {
		log.Fatalln("error when converting max connection: " + err.Error())
	}
	config.MaxConns = int32(maxConnection)

	maxConnectionIdletime, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_IDLETIME"))
	if err != nil {
		log.Fatalln("error when converting max connection idletime: " + err.Error())
	}
	config.MaxConnIdleTime = time.Second * time.Duration(maxConnectionIdletime)

	maxConnectionLifetime, err := strconv.Atoi(os.Getenv("POSTGRES_MAX_LIFETIME"))
	if err != nil {
		log.Fatalln("error when converting max connection lifetime: " + err.Error())
	}
	config.MaxConnLifetime = time.Minute * time.Duration(maxConnectionLifetime)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalln("error when connecting to " + os.Getenv("POSTGRES_HOST") + ", error:" + err.Error())
	}

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalln("error when pinging connection: " + err.Error())
	}

	println(time.Now().String(), "postgres: connected to", os.Getenv("POSTGRES_HOST"))

	return &PostgresUtilImplementation{
		pool: pool,
	}
}

func (util *PostgresUtilImplementation) GetPool() *pgxpool.Pool {
	return util.pool
}

func (util *PostgresUtilImplementation) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	return util.pool.BeginTx(ctx, options)
}

func (util *PostgresUtilImplementation) Close() {
	util.pool.Close()
	println(time.Now().String(), "postgres closed properly")
}

func (util *PostgresUtilImplementation) CommitOrRollback(tx pgx.Tx, ctx context.Context, err error) error {
	if err == nil {
		errCommit := tx.Commit(ctx)
		if errCommit != nil && !errors.Is(errCommit, pgx.ErrTxClosed) {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil && !errors.Is(errRollback, pgx.ErrTxClosed) {
				return errRollback
			}
			return nil
		}
		return nil
	} else {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil && !errors.Is(errRollback, pgx.ErrTxClosed) {
			return errRollback
		}
		return nil
	}
}
