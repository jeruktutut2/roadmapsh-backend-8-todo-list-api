package modelrequests

type CreateTodoRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}
