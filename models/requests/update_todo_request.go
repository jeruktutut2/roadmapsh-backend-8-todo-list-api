package modelrequests

type UpdateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
