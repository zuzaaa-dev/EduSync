package email

import (
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"errors"
	"time"
)

type postgresEmailConfRepo struct {
	db *sql.DB
}

func NewEmailConfirmationsRepository(db *sql.DB) repository.EmailConfirmationsRepository {
	return &postgresEmailConfRepo{db: db}
}

func (r *postgresEmailConfRepo) Create(ctx context.Context, userID int, action, code string, expiresAt time.Time) error {
	// upsert: удаляем старый
	_, err := r.db.ExecContext(ctx, `
        DELETE FROM email_confirmations
         WHERE user_id=$1 AND action=$2
    `, userID, action)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
        INSERT INTO email_confirmations (user_id, action, code, expires_at)
        VALUES ($1, $2, $3, $4)
    `, userID, action, code, expiresAt)
	return err
}

func (r *postgresEmailConfRepo) GetValid(ctx context.Context, userID int, action, code string) (bool, error) {
	var used bool
	var expires time.Time
	err := r.db.QueryRowContext(ctx, `
        SELECT used, expires_at
          FROM email_confirmations
         WHERE user_id=$1 AND action=$2 AND code=$3
    `, userID, action, code).Scan(&used, &expires)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if used || time.Now().After(expires) {
		return false, nil
	}
	return true, nil
}

func (r *postgresEmailConfRepo) MarkUsed(ctx context.Context, userID int, action, code string) error {
	res, err := r.db.ExecContext(ctx, `
        UPDATE email_confirmations
           SET used = TRUE
         WHERE user_id=$1 AND action=$2 AND code=$3
    `, userID, action, code)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("not found")
	}
	return nil
}

// Ограничение по времени повторной отправки
func (r *postgresEmailConfRepo) CanSendNew(ctx context.Context, userID int, action string, throttle time.Duration) (bool, error) {
	var created time.Time
	err := r.db.QueryRowContext(ctx, `
        SELECT created_at 
          FROM email_confirmations
         WHERE user_id=$1 AND action=$2
    `, userID, action).Scan(&created)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return time.Since(created) >= throttle, nil
}
