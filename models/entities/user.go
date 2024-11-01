package modelentities

import "github.com/jackc/pgx/v5/pgtype"

type User struct {
	Id           pgtype.Int4
	Name         pgtype.Text
	Email        pgtype.Text
	Password     pgtype.Text
	RefreshToken pgtype.Text
}
