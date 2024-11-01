package modelresponses

type TodoResponse struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GetTodoResponse struct {
	Data  []TodoResponse `json:"data"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
	Total int            `json:"total"`
}
