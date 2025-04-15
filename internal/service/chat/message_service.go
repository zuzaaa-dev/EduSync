package chat

import (
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"fmt"

	domainChat "EduSync/internal/domain/chat"
	"github.com/sirupsen/logrus"
)

type messageService struct {
	repo repository.MessageRepository
	log  *logrus.Logger
}

func NewMessageService(repo repository.MessageRepository, logger *logrus.Logger) service.MessageService {
	return &messageService{
		repo: repo,
		log:  logger,
	}
}

func (s *messageService) GetMessages(ctx context.Context, chatID, limit, offset int) ([]*domainChat.Message, error) {
	msgs, err := s.repo.GetMessages(ctx, chatID, limit, offset)
	if err != nil {
		s.log.Errorf("Ошибка получения сообщений для чата %d: %v", chatID, err)
		return nil, fmt.Errorf("не удалось получить сообщения")
	}
	return msgs, nil
}

func (s *messageService) SendMessage(ctx context.Context, msg domainChat.Message) (int, error) {
	// Проверяем, что сообщение содержит хотя бы текст или ссылку на группу сообщений
	if msg.Text == nil && msg.MessageGroupID == nil {
		return 0, fmt.Errorf("сообщение должно содержать текст или файлы")
	}
	id, err := s.repo.CreateMessage(ctx, &msg)
	if err != nil {
		s.log.Errorf("Ошибка создания сообщения: %v", err)
		return 0, fmt.Errorf("не удалось создать сообщение")
	}
	return id, nil
}

func (s *messageService) DeleteMessage(ctx context.Context, messageID int, requesterID int) error {
	// Здесь можно добавить проверку прав: например, проверить, является ли requester автором или владельцем чата.
	err := s.repo.DeleteMessage(ctx, messageID)
	if err != nil {
		s.log.Errorf("Ошибка удаления сообщения %d: %v", messageID, err)
		return fmt.Errorf("не удалось удалить сообщение")
	}
	return nil
}

func (s *messageService) ReplyMessage(ctx context.Context, parentMessageID int, msg domainChat.Message) (int, error) {
	// Устанавливаем parent_message_id
	msg.ParentMessageID = &parentMessageID
	id, err := s.repo.CreateMessage(ctx, &msg)
	if err != nil {
		s.log.Errorf("Ошибка создания ответа на сообщение %d: %v", parentMessageID, err)
		return 0, fmt.Errorf("не удалось создать ответ")
	}
	return id, nil
}

func (s *messageService) SearchMessages(ctx context.Context, chatID int, query string, limit, offset int) ([]*domainChat.Message, error) {
	msgs, err := s.repo.SearchMessages(ctx, chatID, query, limit, offset)
	if err != nil {
		s.log.Errorf("Ошибка поиска сообщений в чате %d: %v", chatID, err)
		return nil, fmt.Errorf("не удалось найти сообщения")
	}
	return msgs, nil
}

func (s *messageService) GetMessageFiles(ctx context.Context, messageID int) ([]*domainChat.FileInfo, error) {
	files, err := s.repo.GetMessageFileInfo(ctx, messageID)
	if err != nil {
		s.log.Errorf("Ошибка получения файлов для сообщения %d: %v", messageID, err)
		return nil, fmt.Errorf("не удалось получить информацию о файлах")
	}
	return files, nil
}
