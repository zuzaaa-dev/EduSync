package chat

import (
	domainChat "EduSync/internal/domain/chat"
	"context"
	"database/sql"
	"fmt"
	"time"

	"EduSync/internal/repository"
)

type chatRepository struct {
	db *sql.DB
}

// NewChatRepository возвращает реализацию chat.Repository.
func NewChatRepository(db *sql.DB) repository.ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateChat(ctx context.Context, c *domainChat.Chat) (int, error) {
	var chatID int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO chats (group_id, owner_id, subject_id, join_code, invite_link, created_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`, c.GroupID, c.OwnerID, c.SubjectID, c.JoinCode, c.InviteLink, time.Now()).Scan(&chatID)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания чата: %w", err)
	}
	return chatID, nil
}

func (r *chatRepository) GetChatByID(ctx context.Context, chatID int) (*domainChat.Chat, error) {
	var c domainChat.Chat
	err := r.db.QueryRowContext(ctx, `
		SELECT id, group_id, owner_id, subject_id, join_code, invite_link, created_at
		FROM chats WHERE id = $1
	`, chatID).Scan(&c.ID, &c.GroupID, &c.OwnerID, &c.SubjectID, &c.JoinCode, &c.InviteLink, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения чата: %w", err)
	}
	return &c, nil
}

func (r *chatRepository) DeleteChat(ctx context.Context, chatID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM chats WHERE id = $1`, chatID)
	if err != nil {
		return fmt.Errorf("ошибка удаления чата: %w", err)
	}
	return nil
}

// JoinChat добавляет участника (студента) в чат.
// Предполагается, что таблица student_chats содержит поля student_id и chat_id.
func (r *chatRepository) JoinChat(ctx context.Context, chatID, userID int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO student_chats (student_id, chat_id)
		VALUES ($1, $2)
	`, userID, chatID)
	if err != nil {
		return fmt.Errorf("ошибка присоединения участника к чату: %w", err)
	}
	return nil
}

func (r *chatRepository) UpdateChatInvite(ctx context.Context, chatID int, newJoinCode, newInviteLink string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE chats SET join_code = $1, invite_link = $2 WHERE id = $3
	`, newJoinCode, newInviteLink, chatID)
	if err != nil {
		return fmt.Errorf("ошибка обновления приглашения чата: %w", err)
	}
	return nil
}

// GetParticipants возвращает список участников чата (студенты и, возможно, учителя).
// Здесь мы, например, получаем всех участников из связующей таблицы.
// Если у вас есть другие таблицы для преподавателей, можно объединять результаты (например, через UNION).
func (r *chatRepository) GetParticipants(ctx context.Context, chatID int) ([]*domainChat.Participant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.full_name, u.is_teacher
		FROM users u
		JOIN student_chats sc ON u.id = sc.student_id
		WHERE sc.chat_id = $1
		UNION
		SELECT u.id, u.full_name, u.is_teacher
		FROM users u
		JOIN chats c ON u.id = c.owner_id
		WHERE c.id = $1
	`, chatID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения участников чата: %w", err)
	}
	defer rows.Close()

	var participants []*domainChat.Participant
	for rows.Next() {
		p := new(domainChat.Participant)
		err := rows.Scan(&p.UserID, &p.FullName, &p.IsTeacher)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования участника: %w", err)
		}
		participants = append(participants, p)
	}
	return participants, nil
}

func (r *chatRepository) RemoveParticipant(ctx context.Context, chatID int, userID int) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM student_chats WHERE chat_id = $1 AND student_id = $2
	`, chatID, userID)
	if err != nil {
		return fmt.Errorf("ошибка удаления участника: %w", err)
	}
	return nil
}

func (r *chatRepository) LeaveChat(ctx context.Context, chatID int, userID int) error {
	// То же, что и RemoveParticipant, так как участник уходит сам
	return r.RemoveParticipant(ctx, chatID, userID)
}

// IsParticipant проверяет, является ли пользователь участником данного чата.
// Предполагается, что участники студентов хранятся в таблице student_chats.
func (r *chatRepository) IsParticipant(ctx context.Context, chatID int, userID int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM student_chats WHERE chat_id = $1 AND student_id = $2
		)
	`, chatID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки участия в чате: %w", err)
	}
	if exists {
		return true, nil
	}

	var ownerID int
	err = r.db.QueryRowContext(ctx, `
		SELECT owner_id FROM chats WHERE id = $1
	`, chatID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("ошибка получения владельца чата: %w", err)
	}
	if ownerID == userID {
		return true, nil
	}
	return false, nil
}

// IsOwner проверяет, является ли пользователь владельцем данного чата.
func (r *chatRepository) IsOwner(ctx context.Context, chatID int, userID int) (bool, error) {
	var ownerID int
	err := r.db.QueryRowContext(ctx, `
		SELECT owner_id FROM chats WHERE id = $1
	`, chatID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("ошибка получения владельца чата: %w", err)
	}
	if ownerID == userID {
		return true, nil
	}
	return false, nil
}
