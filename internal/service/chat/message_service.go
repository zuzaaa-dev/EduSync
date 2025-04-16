package chat

import (
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"fmt"
	"strings"

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
	// Проверка, что значения limit и offset корректны
	if limit <= 0 {
		limit = 10 // значение по умолчанию
	}
	if offset < 0 {
		offset = 0
	}
	msgs, err := s.repo.GetMessages(ctx, chatID, limit, offset)
	if err != nil {
		s.log.Errorf("Ошибка получения сообщений для чата %d: %v", chatID, err)
		return nil, fmt.Errorf("не удалось получить сообщения")
	}
	return msgs, nil
}

func (s *messageService) SendMessage(ctx context.Context, msg domainChat.Message) (int, error) {
	// Дополнительная проверка: если текст не nil, должен содержать ненулевое значение после обрезки пробелов
	if msg.Text != nil {
		trimmed := strings.TrimSpace(*msg.Text)
		if len(trimmed) == 0 {
			msg.Text = nil // Если текст пустой, рассматриваем его как nil
		}
	}
	// Проверяем, что сообщение содержит хотя бы текст или информацию о файлах
	if msg.Text == nil && msg.MessageGroupID == nil {
		return 0, fmt.Errorf("сообщение должно содержать текст или файлы")
	}

	// Дополнительно можно проверить, что chat_id и user_id не равны 0:
	if msg.ChatID <= 0 || msg.UserID <= 0 {
		return 0, fmt.Errorf("неверные данные: chat_id или user_id отсутствуют")
	}
	id, err := s.repo.CreateMessage(ctx, &msg)
	if err != nil {
		s.log.Errorf("Ошибка создания сообщения: %v", err)
		return 0, fmt.Errorf("не удалось создать сообщение")
	}
	return id, nil
}

func (s *messageService) DeleteMessage(ctx context.Context, messageID int, requesterID int) error {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		s.log.Errorf("Ошибка получения сообщения %d: %v", messageID, err)
		return fmt.Errorf("не удалось удалить сообщение")
	}
	if msg == nil {
		return fmt.Errorf("сообщение не найдено")
	}
	isTeacher := ctx.Value("is_teacher")
	if requesterID != msg.UserID && !isTeacher.(bool) {
		return fmt.Errorf("нет прав для удаления сообщения")
	}
	err = s.repo.DeleteMessage(ctx, messageID)
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
