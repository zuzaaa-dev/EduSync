package chat

import "time"

// Message — доменная модель сообщения
type Message struct {
	ID              int        `json:"id"`
	ChatID          int        `json:"chat_id"`
	UserID          int        `json:"user_id"`
	Text            *string    `json:"text,omitempty"`
	MessageGroupID  *int       `json:"message_group_id,omitempty"`
	ParentMessageID *int       `json:"parent_message_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	Files           []FileInfo `json:"files,omitempty"`
}

// FileInfo представляет информацию о файле, прикрепленном к сообщению.
type FileInfo struct {
	ID      int    `json:"id"`
	FileURL string `json:"file_url"`
}
