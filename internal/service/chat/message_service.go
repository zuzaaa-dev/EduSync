package chat

import (
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"os"
	"strings"

	domainChat "EduSync/internal/domain/chat"
	"github.com/sirupsen/logrus"
)

var ErrInternal = errors.New("внутренняя ошибка сервера")

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

	// Для каждого сообщения добираем список файлов
	for _, m := range msgs {
		files, ferr := s.repo.GetMessageFileInfo(ctx, m.ID)
		if ferr != nil {
			s.log.Errorf("GetMessageFileInfo(%d): %v", m.ID, ferr)
		} else {
			m.Files = make([]domainChat.FileInfo, len(files))
			for i, f := range files {
				m.Files[i] = *f
			}
		}
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
	tx, err := s.repo.BeginTx(ctx)

	if err != nil {
		s.log.Error("BeginTx:", err)
		return ErrInternal
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1) список файлов (чтобы удалить физически)
	files, ferr := s.repo.GetMessageFileInfo(ctx, messageID)
	if ferr != nil {
		s.log.Error("GetMessageFileInfo:", ferr)
		return ErrInternal
	}

	// 2) удаляем записи в БД
	if err = s.repo.DeleteMessageFilesTx(ctx, tx, messageID); err != nil {
		s.log.Error("DeleteMessageFilesTx:", err)
		return ErrInternal
	}
	if err = s.repo.DeleteMessageTx(ctx, tx, messageID); err != nil {
		s.log.Error("DeleteMessageTx:", err)
		return ErrInternal
	}

	// 3) очищаем физические файлы
	for _, f := range files {
		os.Remove(f.FileURL) // игнорируем ошибку
	}

	if err = tx.Commit(); err != nil {
		s.log.Error("Commit:", err)
		return ErrInternal
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
	for _, m := range msgs {
		files, ferr := s.repo.GetMessageFileInfo(ctx, m.ID)
		if ferr != nil {
			s.log.Errorf("GetMessageFileInfo(%d): %v", m.ID, ferr)
		} else {
			m.Files = make([]domainChat.FileInfo, len(files))
			for i, f := range files {
				m.Files[i] = *f
			}
		}
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

func (s *messageService) SendMessageWithFiles(
	ctx context.Context,
	msg domainChat.Message,
	files []*multipart.FileHeader,
	c *gin.Context,
) (int, error) {
	// Валидация: хотя бы текст или хотя бы один файл
	if msg.Text == nil && len(files) == 0 {
		return 0, errors.New("сообщение должно содержать текст или файл")
	}
	// Начинаем транзакцию через репозиторий
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		s.log.Error("tx begin:", err)
		return 0, errors.New("unexpected error")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1) Создаём сообщение
	msgID, err := s.repo.CreateMessageTx(ctx, tx, &msg)
	if err != nil {
		s.log.Error("create msg:", err)
		return 0, errors.New("failed to save message")
	}

	// 2) Сохраняем файлы (если есть)
	for _, fh := range files {
		// a) сохраняем физически и получаем URL
		dst := fmt.Sprintf("uploads/%d_%s", msgID, fh.Filename)
		if err := c.SaveUploadedFile(fh, dst); err != nil {
			return 0, errors.New("failed to save file")
		}
		// b) сохраняем в БД
		if err := s.repo.CreateMessageFileTx(ctx, tx, msgID, dst); err != nil {
			s.log.Error("save file record:", err)
			return 0, errors.New("failed to save file record")
		}
	}

	if err = tx.Commit(); err != nil {
		s.log.Error("tx commit:", err)
		return 0, errors.New("unexpected error")
	}
	return msgID, nil
}
