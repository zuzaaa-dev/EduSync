package material

import (
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"

	domainChat "EduSync/internal/domain/chat"
)

type fileRepo struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) repository.FileRepository {
	return &fileRepo{db: db}
}

func (r *fileRepo) ByID(ctx context.Context, fileID int) (*domainChat.File, error) {
	const q = `
      SELECT id, message_id, file_url
      FROM message_files
      WHERE id = $1
    `
	f := &domainChat.File{}
	err := r.db.
		QueryRowContext(ctx, q, fileID).
		Scan(&f.ID, &f.MessageID, &f.FileURL)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("file_repository.ByID: %w", err)
	}
	return f, nil
}
