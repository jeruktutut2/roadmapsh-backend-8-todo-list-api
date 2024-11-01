package modelentities

import "github.com/jackc/pgx/v5/pgtype"

type Todo struct {
	Id          pgtype.Int4
	UserId      pgtype.Int4
	Title       pgtype.Text
	Description pgtype.Text
}
