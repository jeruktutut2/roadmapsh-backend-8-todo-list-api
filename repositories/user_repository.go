package repositories

import (
	"context"
	modelentities "todo-list-api/models/entities"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(tx pgx.Tx, ctx context.Context, user modelentities.User) (lastInsertedId int, err error)
	UpdateRefreshToken(tx pgx.Tx, ctx context.Context, refreshToken string, id int) (rowsAffected int64, err error)
	FindByEmail(tx pgx.Tx, ctx context.Context, email string) (user modelentities.User, err error)
	FindByRefreshToken(pool *pgxpool.Pool, ctx context.Context, refreshToken string) (user modelentities.User, err error)
}

type UserRepositoryImplementation struct {
}

func NewUserRepository() UserRepository {
	return &UserRepositoryImplementation{}
}

func (repository *UserRepositoryImplementation) Create(tx pgx.Tx, ctx context.Context, user modelentities.User) (lastInsertedId int, err error) {
	query := `INSERT INTO users (name,email,password) VALUES ($1,$2,$3) RETURNING id;`
	err = tx.QueryRow(ctx, query, user.Name, user.Email, user.Password).Scan(&lastInsertedId)
	return
}

func (repository *UserRepositoryImplementation) UpdateRefreshToken(tx pgx.Tx, ctx context.Context, refreshToken string, id int) (rowsAffected int64, err error) {
	query := `UPDATE users SET refresh_token = $1 WHERE id = $2;`
	result, err := tx.Exec(ctx, query, refreshToken, id)
	if err != nil {
		return
	}
	rowsAffected = result.RowsAffected()
	return
}

func (repository *UserRepositoryImplementation) FindByEmail(tx pgx.Tx, ctx context.Context, email string) (user modelentities.User, err error) {
	query := `SELECT id,name,email,password FROM users WHERE email = $1;`
	err = tx.QueryRow(ctx, query, email).Scan(&user.Id, &user.Name, &user.Email, &user.Password)
	return
}

func (repository *UserRepositoryImplementation) FindByRefreshToken(pool *pgxpool.Pool, ctx context.Context, refreshToken string) (user modelentities.User, err error) {
	query := `SELECT id,name,email,password,refresh_token FROM users WHERE refresh_token = $1;`
	err = pool.QueryRow(ctx, query, refreshToken).Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.RefreshToken)
	return
}
