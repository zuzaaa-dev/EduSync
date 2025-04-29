package favorite

import (
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
)

type fileFavoriteRepo struct {
	db *sql.DB
}

func NewFileFavoriteRepository(db *sql.DB) repository.FileFavoriteRepository {
	return &fileFavoriteRepo{db: db}
}

func (r *fileFavoriteRepo) Add(ctx context.Context, userID, fileID int) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO favorites (user_id, file_id)
        VALUES ($1, $2)
    `, userID, fileID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return ErrAlreadyFavorited
		}
		return fmt.Errorf("fileFavoriteRepo.Add: %w", err)
	}
	return nil
}

func (r *fileFavoriteRepo) Remove(ctx context.Context, userID, fileID int) error {
	res, err := r.db.ExecContext(ctx, `
        DELETE FROM favorites WHERE user_id = $1 AND file_id = $2
    `, userID, fileID)
	if err != nil {
		return fmt.Errorf("fileFavoriteRepo.Remove: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrFavoriteNotFound
	}
	return nil
}

func (r *fileFavoriteRepo) Exists(ctx context.Context, userID, fileID int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
        SELECT true FROM favorites WHERE user_id=$1 AND file_id=$2
    `, userID, fileID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("fileFavoriteRepo.Exists: %w", err)
	}
	return true, nil
}

func (r *fileFavoriteRepo) ListByUser(ctx context.Context, userID int) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT file_id
          FROM favorites
         WHERE user_id = $1
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("fileFavoriteRepo.ListByUser: %w", err)
	}
	defer rows.Close()

	var out []int
	for rows.Next() {
		var fid int
		if err := rows.Scan(&fid); err != nil {
			return nil, fmt.Errorf("fileFavoriteRepo.ListByUser scan: %w", err)
		}
		out = append(out, fid)
	}
	return out, nil
}
