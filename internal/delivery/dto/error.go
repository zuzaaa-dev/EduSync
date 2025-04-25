package dto

// ErrorResponse тело запроса обновления токена
// swagger:response ErrorResponse
type ErrorResponse struct {
	// in:body
	Body struct {
		// Сообщение об ошибке
		// example: "Неверный формат запроса"
		Error string `json:"error"`
	}
}
