package chat

import "time"

// Chat представляет доменную модель чата.
type Chat struct {
	ID         int       `json:"id"`
	GroupID    int       `json:"group_id"`
	OwnerID    int       `json:"owner_id"`
	SubjectID  int       `json:"subject_id"`
	JoinCode   string    `json:"join_code"`
	InviteLink string    `json:"invite_link"`
	CreatedAt  time.Time `json:"created_at"`
}

// Participant представляет пользователя, участника чата.
type Participant struct {
	UserID    int    `json:"user_id"`
	FullName  string `json:"full_name"`
	IsTeacher bool   `json:"is_teacher"`
}

// Message представляет сообщение в чате.
type Message struct {
	ID              int       `json:"id"`
	ChatID          int       `json:"chat_id"`
	UserID          int       `json:"user_id"`
	Text            *string   `json:"text"`              // может быть nil, если только файлы
	MessageGroupID  *int      `json:"message_group_id"`  // для группировки сообщений (если применимо)
	ParentMessageID *int      `json:"parent_message_id"` // для ответов
	CreatedAt       time.Time `json:"created_at"`
}

// FileInfo представляет информацию о файле, прикрепленном к сообщению.
type FileInfo struct {
	ID      int    `json:"id"`
	FileURL string `json:"file_url"`
}
