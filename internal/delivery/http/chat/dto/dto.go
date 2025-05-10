package dto

import "time"

// CreatePollReq — тело запроса на создание опроса.
// swagger:model CreatePollReq
type CreatePollReq struct {
	// in:body
	// Вопрос опроса
	// example: "Как вам новый формат занятий?"
	Question string `json:"question" binding:"required,min=1,max=255"`
	// Варианты ответа (минимум 2)
	// example: ["Очень понравилось", "Нормально", "Не понравилось"]
	Options []string `json:"options" binding:"required,min=1,max=8,dive,required,min=1,max=255"`
}

// VoteReq — тело запроса на голос.
// swagger:model VoteReq
type VoteReq struct {
	// ID выбранного варианта
	// example: 3
	OptionID int `json:"option_id" binding:"required"`
}

type PollSummary struct {
	ID        int               `json:"id"`
	Question  string            `json:"question"`
	CreatedAt time.Time         `json:"created_at"`
	Options   []OptionWithCount `json:"options"`
}
type OptionWithCount struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	Votes int    `json:"votes"`
}
