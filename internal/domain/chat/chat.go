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
