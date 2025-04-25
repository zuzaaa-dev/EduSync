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

	// ID предмета
	// example: 3
	SubjectID int `json:"subject_id"`

	// Код присоединения
	// example: ABC123
	JoinCode string `json:"join_code"`

	// Ссылка-приглашение
	// example: https://edusync.com/join/ABC123
	InviteLink string `json:"invite_link"`
}

func ConvertChatToDTO(chat *domainChat.Chat) *ChatInfo {
	return &ChatInfo{
		ID:         chat.ID,
		GroupID:    chat.GroupID,
		OwnerID:    chat.OwnerID,
		SubjectID:  chat.SubjectID,
		JoinCode:   chat.JoinCode,
		InviteLink: chat.InviteLink,
	}
}
