package helpers

type Response struct {
	Message string `json:"message"`
}

func ToResponse(message string) Response {
	return Response{
		Message: message,
	}
}
