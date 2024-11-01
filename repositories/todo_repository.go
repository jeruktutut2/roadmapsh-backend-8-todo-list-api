package repositories

import (
	"context"
	modelentities "todo-list-api/models/entities"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoRepository interface {
	Create(tx pgx.Tx, ctx context.Context, todo modelentities.Todo) (lastInsertId int, err error)
	FindByIdAndUserId(tx pgx.Tx, ctx context.Context, id int, userId int) (todo modelentities.Todo, err error)
	Update(tx pgx.Tx, ctx context.Context, todo modelentities.Todo) (rowsAffected int64, err error)
	Delete(tx pgx.Tx, ctx context.Context, id int) (rowsAffected int64, err error)
	FindByPagination(pool *pgxpool.Pool, ctx context.Context, userId int, offset int, limit int) (todos []modelentities.Todo, err error)
	Count(pool *pgxpool.Pool, ctx context.Context) (numberOfTodos int, err error)
}

type TodoRepositoryImplementation struct {
}

func NewTodoRepository() TodoRepository {
	return &TodoRepositoryImplementation{}
}

func (repository *TodoRepositoryImplementation) Create(tx pgx.Tx, ctx context.Context, todo modelentities.Todo) (lastInsertId int, err error) {
	query := `INSERT INTO todos (user_id,title,description) VALUES ($1,$2,$3) RETURNING id;`
	err = tx.QueryRow(ctx, query, todo.UserId, todo.Title, todo.Description).Scan(&lastInsertId)
	return
}

func (repository *TodoRepositoryImplementation) FindByIdAndUserId(tx pgx.Tx, ctx context.Context, id int, userId int) (todo modelentities.Todo, err error) {
	query := `SELECT id, user_id, title, description FROM todos WHERE id = $1 AND user_id = $2;`
	err = tx.QueryRow(ctx, query, id, userId).Scan(&todo.Id, &todo.UserId, &todo.Title, &todo.Description)
	return
}

func (repository *TodoRepositoryImplementation) Update(tx pgx.Tx, ctx context.Context, todo modelentities.Todo) (rowsAffected int64, err error) {
	query := `UPDATE todos SET title = $1, description = $2 WHERE id = $3;`
	result, err := tx.Exec(ctx, query, todo.Title, todo.Description, todo.Id)
	if err != nil {
		return
	}
	rowsAffected = result.RowsAffected()
	return
}

func (repository *TodoRepositoryImplementation) Delete(tx pgx.Tx, ctx context.Context, id int) (rowsAffected int64, err error) {
	query := `DELETE FROM todos WHERE id = $1;`
	result, err := tx.Exec(ctx, query, id)
	if err != nil {
		return
	}
	rowsAffected = result.RowsAffected()
	return
}

func (repository *TodoRepositoryImplementation) FindByPagination(pool *pgxpool.Pool, ctx context.Context, userId int, offset int, limit int) (todos []modelentities.Todo, err error) {
	query := `SELECT id, user_id, title, description FROM todos WHERE user_id = $1 ORDER BY id ASC OFFSET $2 LIMIT $3;`
	rows, err := pool.Query(ctx, query, userId, offset, limit)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var todo modelentities.Todo
		err = rows.Scan(&todo.Id, &todo.UserId, &todo.Title, &todo.Description)
		if err != nil {
			todos = []modelentities.Todo{}
			return
		}
		todos = append(todos, todo)
	}

	if rows.Err() != nil {
		todos = []modelentities.Todo{}
		err = rows.Err()
		return
	}
	return
}

func (repository *TodoRepositoryImplementation) Count(pool *pgxpool.Pool, ctx context.Context) (numberOfTodos int, err error) {
	query := `SELECT COUNT(*) AS number_of_todos FROM todos;`
	err = pool.QueryRow(ctx, query).Scan(&numberOfTodos)
	return
}
