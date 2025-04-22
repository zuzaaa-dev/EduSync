package chat

import "time"

// Chat модель чата
// swagger:model
type Chat struct {
	// ID чата
	// example: 1
	ID int `json:"id"`

	// ID группы
	// example: 5
	GroupID int `json:"group_id"`

	// ID владельца
	// example: 10
	OwnerID int `json:"owner_id"`

	// ID предмета
	// example: 3
	SubjectID int `json:"subject_id"`

	// Код присоединения
	// example: ABC123
	JoinCode string `json:"join_code"`

	// Ссылка-приглашение
	// example: https://edusync.com/join/ABC123
	InviteLink string `json:"invite_link"`

	// Дата создания
	// example: 2023-01-15T09:30:00Z
	CreatedAt time.Time `json:"created_at"`
}

// Participant модель участника
// swagger:model
type Participant struct {
	// ID пользователя
	// example: 7
	UserID int `json:"user_id"`

	// Полное имя
	// example: Иван Иванов
	FullName string `json:"full_name"`

	// Статус преподавателя
	// example: false
	IsTeacher bool `json:"is_teacher"`
}
