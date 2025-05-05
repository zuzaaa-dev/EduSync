package chat

import (
	"EduSync/internal/delivery/ws"
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"os"
	"strings"
	"time"

	domainChat "EduSync/internal/domain/chat"
	"github.com/sirupsen/logrus"
)

var ErrInternal = errors.New("внутренняя ошибка сервера")

type messageService struct {
	repo repository.MessageRepository
	log  *logrus.Logger
	hub  *ws.Hub
}

func NewMessageService(repo repository.MessageRepository, logger *logrus.Logger, hub *ws.Hub) service.MessageService {
	return &messageService{
		repo: repo,
		log:  logger,
		hub:  hub,
	}
}

func (s *messageService) Messages(ctx context.Context, chatID, limit, offset int) ([]*domainChat.Message, error) {
	// Проверка, что значения limit и offset корректны
	if limit <= 0 {
		limit = 10 // значение по умолчанию
	}
	if offset < 0 {
		offset = 0
	}
	msgs, err := s.repo.Messages(ctx, chatID, limit, offset)
	if err != nil {
		s.log.Errorf("Ошибка получения сообщений для чата %d: %v", chatID, err)
		return nil, fmt.Errorf("не удалось получить сообщения")
	}

	// Для каждого сообщения добираем список файлов
	for _, m := range msgs {
		files, ferr := s.repo.MessageFileInfo(ctx, m.ID)
		if ferr != nil {
			s.log.Errorf("MessageFileInfo(%d): %v", m.ID, ferr)
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

	// Загружаем только что созданное сообщение целиком (со всеми файлами), чтобы отдать в WS
	created, err := s.repo.ByID(ctx, id)
	if err != nil {
		s.log.Errorf("ошибка получения нового сообщения %d: %v", id, err)
	} else if created != nil {
		// broadcast в комнату chat_<chatID>
		room := fmt.Sprintf("chat_%d", created.ChatID)
		s.hub.Broadcast(room, "message:new", created)
	}

	return id, nil
}

func (s *messageService) DeleteMessage(ctx context.Context, messageID int, requesterID int) error {
	msg, err := s.repo.ByID(ctx, messageID)
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
	files, ferr := s.repo.MessageFileInfo(ctx, messageID)
	if ferr != nil {
		s.log.Error("MessageFileInfo:", ferr)
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
		return ErrInternal
	}

	// после успешного удаления
	room := fmt.Sprintf("chat_%d", msg.ChatID)
	s.hub.Broadcast(room, "message:delete", map[string]int{"id": messageID})

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
	reply, err := s.repo.ByID(ctx, id)
	if err == nil && reply != nil {
		room := fmt.Sprintf("chat_%d", reply.ChatID)
		s.hub.Broadcast(room, "message:new", reply)
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
		files, ferr := s.repo.MessageFileInfo(ctx, m.ID)
		if ferr != nil {
			s.log.Errorf("MessageFileInfo(%d): %v", m.ID, ferr)
		} else {
			m.Files = make([]domainChat.FileInfo, len(files))
			for i, f := range files {
				m.Files[i] = *f
			}
		}
	}
	return msgs, nil
}

func (s *messageService) MessageFiles(ctx context.Context, messageID int) ([]*domainChat.FileInfo, error) {
	files, err := s.repo.MessageFileInfo(ctx, messageID)
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
	// Соберем FileInfo для отправки клиентам
	var attachedFiles []domainChat.FileInfo

	for _, fh := range files {
		// a) сохраняем физически и получаем URL
		dst := fmt.Sprintf("uploads/%d_%s", msgID, fh.Filename)
		if err := c.SaveUploadedFile(fh, dst); err != nil {
			return 0, errors.New("failed to save file")
		}
		// b) сохраняем в БД
		fileID, err := s.repo.CreateMessageFileTx(ctx, tx, msgID, dst)
		if err != nil {
			s.log.Error("save file record:", err)
			return 0, errors.New("failed to save file record")
		}
		attachedFiles = append(attachedFiles, domainChat.FileInfo{
			ID:      fileID,
			FileURL: dst,
		})
	}

	if err = tx.Commit(); err != nil {
		s.log.Error("tx commit:", err)
		return 0, errors.New("unexpected error")
	}
	outgoing := domainChat.Message{
		ID:              msgID,
		ChatID:          msg.ChatID,
		UserID:          msg.UserID,
		Text:            msg.Text,
		MessageGroupID:  msg.MessageGroupID,
		ParentMessageID: msg.ParentMessageID,
		CreatedAt:       time.Now(),
		Files:           attachedFiles,
	}
	room := fmt.Sprintf("chat_%d", msg.ChatID)
	s.hub.Broadcast(room, "message:new", outgoing)
	return msgID, nil
}

func (s *messageService) UpdateMessage(
	ctx context.Context,
	messageID, requesterID int,
	newText *string,
) (*domainChat.Message, error) {
	// 1) Достать текущее сообщение
	orig, err := s.repo.ByID(ctx, messageID)
	if err != nil {
		s.log.Errorf("UpdateMessage: fetch orig: %v", err)
		return nil, ErrInternal
	}
	if orig == nil {
		return nil, domainChat.ErrNotFound
	}
	// 2) Проверить права: автор или преподаватель
	isTeacher := ctx.Value("is_teacher").(bool)
	if orig.UserID != requesterID && !isTeacher {
		return nil, domainChat.ErrPermissionDenied
	}
	// 3) Трим и валидация
	text := ""
	if newText != nil {
		text = strings.TrimSpace(*newText)
		if text == "" {
			// разрешим обнулить текст?
			return nil, fmt.Errorf("текст не может быть пустым")
		}
	} else {
		return nil, fmt.Errorf("нет поля text для обновления")
	}
	// 4) Начать транзакцию
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		s.log.Errorf("UpdateMessage: begin tx: %v", err)
		return nil, ErrInternal
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	// 5) Обновить
	if err = s.repo.UpdateMessageTx(ctx, tx, messageID, text); err != nil {
		s.log.Errorf("UpdateMessage: exec update: %v", err)
		return nil, ErrInternal
	}
	if err = tx.Commit(); err != nil {
		s.log.Errorf("UpdateMessage: commit: %v", err)
		return nil, ErrInternal
	}
	// 6) Подготовить объект для отдачи и WS
	updated, err := s.repo.ByID(ctx, messageID)
	if err != nil {
		s.log.Errorf("UpdateMessage: fetch updated: %v", err)
		return nil, ErrInternal
	}
	// 7) Рассылка в WS
	room := fmt.Sprintf("chat_%d", updated.ChatID)
	s.hub.Broadcast(room, "message:updated", updated)
	return updated, nil
}
