package chat

import (
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"time"

	domainChat "EduSync/internal/domain/chat"
)

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) repository.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) CreateMessage(ctx context.Context, msg *domainChat.Message) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO messages (chat_id, user_id, text, message_group_id, parent_message_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`, msg.ChatID, msg.UserID, msg.Text, msg.MessageGroupID, msg.ParentMessageID, time.Now()).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания сообщения: %w", err)
	}
	return id, nil
}

func (r *messageRepository) GetMessageByID(ctx context.Context, msgID int) (*domainChat.Message, error) {
	msg := &domainChat.Message{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, chat_id, user_id, text, message_group_id, parent_message_id, created_at
		FROM messages
		WHERE id = $1`, msgID).Scan(&msg.ID, &msg.ChatID, &msg.UserID, &msg.Text, &msg.MessageGroupID, &msg.ParentMessageID, &msg.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сообщенияg по id: %w", err)
	}
	return msg, nil
}

func (r *messageRepository) GetMessages(ctx context.Context, chatID int, limit, offset int) ([]*domainChat.Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, chat_id, user_id, text, message_group_id, parent_message_id, created_at 
		FROM messages 
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сообщений: %w", err)
	}
	defer rows.Close()

	var messages []*domainChat.Message
	for rows.Next() {
		msg := new(domainChat.Message)
		err := rows.Scan(&msg.ID, &msg.ChatID, &msg.UserID, &msg.Text, &msg.MessageGroupID, &msg.ParentMessageID, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования сообщения: %w", err)
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (r *messageRepository) DeleteMessage(ctx context.Context, messageID int) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM messages WHERE id = $1
	`, messageID)
	if err != nil {
		return fmt.Errorf("ошибка удаления сообщения: %w", err)
	}
	return nil
}

func (r *messageRepository) SearchMessages(ctx context.Context, chatID int, query string, limit, offset int) ([]*domainChat.Message, error) {
	// Ищем и по тексту, и по file_url (чтобы нашли по имени файла)
	rows, err := r.db.QueryContext(ctx, `
    SELECT DISTINCT m.id, m.chat_id, m.user_id, m.text, m.message_group_id, m.parent_message_id, m.created_at
      FROM messages m
      LEFT JOIN message_files mf ON mf.message_id = m.id
     WHERE m.chat_id = $1
       AND (m.text ILIKE '%'||$2||'%' OR mf.file_url ILIKE '%'||$2||'%')
     ORDER BY m.created_at
     LIMIT $3 OFFSET $4
  `, chatID, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("SearchMessages query: %w", err)
	}
	defer rows.Close()

	var msgs []*domainChat.Message
	for rows.Next() {
		m := &domainChat.Message{}
		if err := rows.Scan(&m.ID, &m.ChatID, &m.UserID, &m.Text, &m.MessageGroupID, &m.ParentMessageID, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("SearchMessages scan: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *messageRepository) GetMessageFileInfo(ctx context.Context, messageID int) ([]*domainChat.FileInfo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, file_url FROM message_files WHERE message_id = $1
	`, messageID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации о файлах: %w", err)
	}
	defer rows.Close()

	var files []*domainChat.FileInfo
	for rows.Next() {
		file := new(domainChat.FileInfo)
		if err := rows.Scan(&file.ID, &file.FileURL); err != nil {
			return nil, fmt.Errorf("ошибка сканирования информации о файле: %w", err)
		}
		files = append(files, file)
	}
	return files, nil
}

func (r *messageRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *messageRepository) CreateMessageTx(
	ctx context.Context, tx *sql.Tx, msg *domainChat.Message,
) (int, error) {
	var id int
	err := tx.QueryRowContext(ctx, `
        INSERT INTO messages
          (chat_id, user_id, text, message_group_id, parent_message_id, created_at)
        VALUES ($1,$2,$3,$4,$5,$6) RETURNING id
    `, msg.ChatID, msg.UserID, msg.Text, msg.MessageGroupID, msg.ParentMessageID, time.Now()).
		Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *messageRepository) CreateMessageFileTx(
	ctx context.Context, tx *sql.Tx, messageID int, fileURL string,
) error {
	_, err := tx.ExecContext(ctx, `
        INSERT INTO message_files (message_id, file_url)
        VALUES ($1,$2)
    `, messageID, fileURL)
	return err
}

func (r *messageRepository) DeleteMessageFilesTx(ctx context.Context, tx *sql.Tx, messageID int) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM message_files WHERE message_id = $1`, messageID)
	return err
}

func (r *messageRepository) DeleteMessageTx(ctx context.Context, tx *sql.Tx, messageID int) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM messages WHERE id = $1`, messageID)
	return err
}
