package chat

import (
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	dtoChat "EduSync/internal/delivery/dto/chat"
	domainChat "EduSync/internal/domain/chat"

	"github.com/sirupsen/logrus"
)

type chatService struct {
	repo     repository.ChatRepository
	subjRepo repository.SubjectRepository
	userRepo repository.UserRepository
	log      *logrus.Logger
}

func NewChatService(
	repo repository.ChatRepository,
	subjRepo repository.SubjectRepository,
	userRepo repository.UserRepository,
	log *logrus.Logger,
) service.ChatService {
	return &chatService{
		repo:     repo,
		subjRepo: subjRepo,
		userRepo: userRepo,
		log:      log,
	}
}

// CreateChat создает чат и генерирует приглашения.
func (s *chatService) CreateChat(ctx context.Context, c domainChat.Chat) (*domainChat.Chat, error) {
	c.JoinCode = generateRandomCode(10)
	c.InviteLink = fmt.Sprintf("https://edusync.ru/invite/%s", c.JoinCode)
	c.CreatedAt = time.Now()
	chat, err := s.repo.CreateChat(ctx, &c)
	if err != nil {
		s.log.Errorf("Ошибка создания чата: %v", err)
		return nil, fmt.Errorf("не удалось создать чат")
	}
	s.log.Infof("Чат создан с ID: %d", chat.ID)

	return chat, nil
}

// ListForUser возвращает все чаты, в которых участвует userID или которые он создал.
func (s *chatService) ListForUser(ctx context.Context, userID int, isTeacher bool) ([]*dtoChat.ChatInfo, error) {
	s.log.Infof("ListForUser user=%d teacher=%v", userID, isTeacher)
	chats, err := s.repo.ForUser(ctx, userID, isTeacher)
	if err != nil {
		s.log.Errorf("repo.ForUser: %v", err)
		return nil, fmt.Errorf("не удалось получить список чатов")
	}

	out := make([]*dtoChat.ChatInfo, 0, len(chats))
	for _, c := range chats {
		info, err := s.buildChatInfo(ctx, c)
		if err != nil {
			s.log.Errorf("buildChatInfo %+v: %v", c, err)
			continue
		}
		out = append(out, info)
	}
	return out, nil
}

// ChatParticipants возвращает список участников чата.
func (s *chatService) ChatParticipants(ctx context.Context, chatID int) ([]*domainChat.Participant, error) {
	participants, err := s.repo.GetParticipants(ctx, chatID)
	if err != nil {
		s.log.Errorf("Ошибка получения участников для чата %d: %v", chatID, err)
		return nil, fmt.Errorf("не удалось получить список участников")
	}
	return participants, nil
}

// JoinChat присоединяет userID к чату и возвращает обновлённую информацию о нём.
func (s *chatService) JoinChat(ctx context.Context, userID int, code string) (*dtoChat.ChatInfo, error) {
	c, err := s.repo.ChatByCode(ctx, code)
	if err != nil {
		s.log.Errorf("repo.ChatByCode(%d): %v", code, err)
		return nil, fmt.Errorf("не удалось присоединиться к чату")
	}
	if c == nil {
		return nil, fmt.Errorf("чат не найден")
	}

	// валидация кода
	if code != c.JoinCode && code != c.InviteLink {
		return nil, domainChat.ErrInvalidJoinCode
	}

	if err := s.repo.JoinChat(ctx, c.ID, userID); err != nil {
		s.log.Errorf("repo.JoinChat(%d,%d): %v", c.ID, userID, err)
		return nil, fmt.Errorf("не удалось присоединиться к чату")
	}

	// вернём DTO
	info, err := s.buildChatInfo(ctx, c)
	if err != nil {
		s.log.Errorf("buildChatInfo after join: %v", err)
		return nil, fmt.Errorf("не удалось получить информацию о чате")
	}
	return info, nil
}

// RecreateInvite генерирует новый join_code и invite_link, если запрос делает владелец.
func (s *chatService) RecreateInvite(ctx context.Context, chatID int, ownerID int) (*domainChat.Chat, error) {
	chat, err := s.repo.ChatByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения чата: %w", err)
	}
	if chat == nil || chat.OwnerID != ownerID {
		return nil, fmt.Errorf("нет прав на изменение приглашения чата")
	}
	newCode := generateRandomCode(10)
	newLink := fmt.Sprintf("https://edusync.ru/invite/%s", newCode)
	if err := s.repo.UpdateChatInvite(ctx, chatID, newCode, newLink); err != nil {
		s.log.Errorf("Ошибка обновления приглашения: %v", err)
		return nil, fmt.Errorf("не удалось обновить приглашение")
	}
	chat.InviteLink = newLink
	chat.JoinCode = newCode
	s.log.Infof("Приглашение обновлено для чата ID: %d", chatID)
	return chat, nil
}

// DeleteChat удаляет чат, если запрос отправлен владельцем.
func (s *chatService) DeleteChat(ctx context.Context, chatID int, ownerID int) error {
	chat, err := s.repo.ChatByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("ошибка получения чата: %w", err)
	}
	if chat == nil || chat.OwnerID != ownerID {
		return fmt.Errorf("нет прав на удаление чата")
	}
	if err := s.repo.DeleteChat(ctx, chatID); err != nil {
		s.log.Errorf("Ошибка удаления чата %d: %v", chatID, err)
		return fmt.Errorf("не удалось удалить чат")
	}
	s.log.Infof("Чат %d успешно удален", chatID)
	return nil
}

// RemoveParticipant удаляет участника (только по запросу владельца).
func (s *chatService) RemoveParticipant(ctx context.Context, chatID int, ownerID int, participantID int) error {
	chat, err := s.repo.ChatByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("ошибка получения чата: %w", err)
	}
	if chat == nil || chat.OwnerID != ownerID {
		return fmt.Errorf("нет прав на удаление участника")
	}
	if err := s.repo.RemoveParticipant(ctx, chatID, participantID); err != nil {
		s.log.Errorf("Ошибка удаления участника %d из чата %d: %v", participantID, chatID, err)
		return fmt.Errorf("не удалось удалить участника")
	}
	s.log.Infof("Участник %d удален из чата %d", participantID, chatID)
	return nil
}

// LeaveChat позволяет участнику покинуть чат.
func (s *chatService) LeaveChat(ctx context.Context, chatID int, userID int) error {
	if err := s.repo.LeaveChat(ctx, chatID, userID); err != nil {
		s.log.Errorf("Ошибка, когда пользователь %d покидает чат %d: %v", userID, chatID, err)
		return fmt.Errorf("не удалось покинуть чат")
	}
	s.log.Infof("Пользователь %d покинул чат %d", userID, chatID)
	return nil
}

// generateRandomCode генерирует случайный код указанной длины.
func generateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var code strings.Builder
	// Можно использовать math/rand или crypto/rand для большей безопасности.
	for i := 0; i < length; i++ {
		code.WriteByte(charset[rand.Intn(length)])
	}
	return code.String()
}

// buildChatInfo собирает ChatInfo из domain.Chat, добирая названия и ФИО.
func (s *chatService) buildChatInfo(ctx context.Context, c *domainChat.Chat) (*dtoChat.ChatInfo, error) {
	// 1) subject name
	subj, err := s.subjRepo.ByID(ctx, c.SubjectID)
	if err != nil {
		return nil, fmt.Errorf("subjectRepo.ByID: %w", err)
	}
	subjName := ""
	if subj != nil {
		subjName = subj.Name
	}

	// 2) owner full name
	usr, err := s.userRepo.ByID(ctx, c.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetByID: %w", err)
	}
	ownerName := ""
	if usr != nil {
		ownerName = usr.FullName
	}

	return &dtoChat.ChatInfo{
		ID:            c.ID,
		GroupID:       c.GroupID,
		SubjectID:     c.SubjectID,
		SubjectName:   subjName,
		OwnerID:       c.OwnerID,
		OwnerFullName: ownerName,
	}, nil
}
