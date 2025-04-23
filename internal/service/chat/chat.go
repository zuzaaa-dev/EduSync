package chat

import (
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"fmt"
	"strings"
	"time"

	domainChat "EduSync/internal/domain/chat"
	"github.com/sirupsen/logrus"
)

type chatService struct {
	repo repository.ChatRepository
	log  *logrus.Logger
}

func NewChatService(repo repository.ChatRepository, logger *logrus.Logger) service.ChatService {
	return &chatService{
		repo: repo,
		log:  logger,
	}
}

// CreateChat создает чат и генерирует приглашения.
func (s *chatService) CreateChat(ctx context.Context, c domainChat.Chat) (int, error) {
	// Проверка, что владелец - учитель, и другие бизнес-правила можно добавить здесь.
	// Генерация join_code и invite_link:
	c.JoinCode = generateRandomCode(6)
	c.InviteLink = fmt.Sprintf("https://yourapp.com/invite/%s", c.JoinCode)
	c.CreatedAt = time.Now()
	chatID, err := s.repo.CreateChat(ctx, &c)
	if err != nil {
		s.log.Errorf("Ошибка создания чата: %v", err)
		return 0, fmt.Errorf("не удалось создать чат")
	}
	s.log.Infof("Чат создан с ID: %d", chatID)
	return chatID, nil
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

func (s *chatService) JoinChat(ctx context.Context, chatID, userID int) error {
	// TODO сделать проверку ввода приглос ссылки или кода
	err := s.repo.JoinChat(ctx, chatID, userID)
	if err != nil {
		s.log.Errorf("Ошибка присоединения пользователя %d к чату %d: %v", userID, chatID, err)
		return fmt.Errorf("не удалось присоединиться к чату")
	}
	s.log.Infof("Пользователь %d успешно присоединился к чату %d", userID, chatID)
	return nil
}

// RecreateInvite генерирует новый join_code и invite_link, если запрос делает владелец.
func (s *chatService) RecreateInvite(ctx context.Context, chatID int, ownerID int) error {
	// Проверить, что запрашивающий является владельцем чата (это можно проверить через GetChatByID)
	// TODO добавить возврат новой пригласительной ссылки
	chat, err := s.repo.GetChatByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("ошибка получения чата: %w", err)
	}
	if chat == nil || chat.OwnerID != ownerID {
		return fmt.Errorf("нет прав на изменение приглашения чата")
	}
	newCode := generateRandomCode(10)
	newLink := fmt.Sprintf("https://yourapp.com/invite/%s", newCode)
	if err := s.repo.UpdateChatInvite(ctx, chatID, newCode, newLink); err != nil {
		s.log.Errorf("Ошибка обновления приглашения: %v", err)
		return fmt.Errorf("не удалось обновить приглашение")
	}
	s.log.Infof("Приглашение обновлено для чата ID: %d", chatID)
	return nil
}

// DeleteChat удаляет чат, если запрос отправлен владельцем.
func (s *chatService) DeleteChat(ctx context.Context, chatID int, ownerID int) error {
	chat, err := s.repo.GetChatByID(ctx, chatID)
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
	chat, err := s.repo.GetChatByID(ctx, chatID)
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
		code.WriteByte(charset[time.Now().UnixNano()%int64(len(charset))])
	}
	return code.String()
}
