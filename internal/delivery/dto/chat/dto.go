package chat

import domainChat "EduSync/internal/domain/chat"

// CreateChatRequest модель создания чата
// swagger:model
var CreateChatRequest struct {
	// ID чата
	// example: 1
	GroupID int `json:"group_id" binding:"required"`

	// ID предмета
	// example: 3
	SubjectID int `json:"subject_id" binding:"required"`
}

// ChatInfo модель чата
// swagger:model
type ChatInfo struct {
	// ID чата
	// example: 1
	ID int `json:"id"`

	// ID группы
	// example: 5
	GroupID int `json:"group_id"`

	// ID владельца
	// example: 10
	OwnerID int `json:"owner_id"`

	// Полное имя владельца чата
	// example: Фамилия Имя Отчество
	OwnerFullName string `json:"owner_full_name"`

	// ID предмета
	// example: 3
	SubjectID int `json:"subject_id"`

	// Название предмета
	// example: УП
	SubjectName string `json:"subject_name"`
}

func ConvertChatToDTO(chat *domainChat.Chat) *ChatInfo {
	return &ChatInfo{
		ID:      chat.ID,
		GroupID: chat.GroupID,
		OwnerID: chat.OwnerID,
	}
}
