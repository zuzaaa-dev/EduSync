package chat

import "time"

// swagger:model Poll
type Poll struct {
	// ID опроса
	// example: 7
	ID int `json:"id"`
	// ID чата
	// example: 12
	ChatID int `json:"chat_id"`
	// Вопрос
	// example: "Как вам новый формат занятий?"
	Question string `json:"question"`
	// Когда создан
	CreatedAt time.Time `json:"created_at"`
}

// swagger:model Option
type Option struct {
	// ID варианта
	// example: 3
	ID int `json:"id"`
	// ID опроса
	// example: 7
	PollID int `json:"poll_id"`
	// Текст варианта
	// example: "Очень понравилось"
	Text string `json:"option_text"`
}

// swagger:model Vote
type Vote struct {
	// ID пользователя
	// example: 42
	UserID int `json:"user_id"`
	// ID варианта
	// example: 3
	PollOptionID int `json:"poll_option_id"`
}
