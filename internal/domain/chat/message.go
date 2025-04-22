package chat

import "time"

// Message модель сообщения
// swagger:model
type Message struct {
	// ID сообщения
	// example: 1
	ID int `json:"id"`

	// ID чата
	// example: 5
	ChatID int `json:"chat_id"`

	// ID пользователя
	// example: 10
	UserID int `json:"user_id"`

	// Текст сообщения
	// example: Привет! Как дела?
	Text *string `json:"text,omitempty"`

	// ID группы сообщений
	// example: 3
	MessageGroupID *int `json:"message_group_id,omitempty"`

	// ID родительского сообщения
	// example: 7
	ParentMessageID *int `json:"parent_message_id,omitempty"`

	// Время создания
	// example: 2023-01-15T09:30:00Z
	CreatedAt time.Time `json:"created_at"`

	// Прикрепленные файлы
	Files []FileInfo `json:"files,omitempty"`
}

// FileInfo модель файла
// swagger:model
type FileInfo struct {
	// ID файла
	// example: 1
	ID int `json:"id"`

	// Ссылка на файл
	// example: https://storage.example.com/files/abc123.pdf
	FileURL string `json:"file_url"`
}
