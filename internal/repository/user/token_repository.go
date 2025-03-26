package user

import (
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// TokenRepository обеспечивает работу с таблицей токенов.
type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) repository.TokenRepository {
	return &TokenRepository{db: db}
}

// DeleteTokensForUser удаляет все токены для пользователя.
// Это позволит гарантировать, что при новом логине не будет дублирования.
func (r *TokenRepository) DeleteTokensForUser(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tokens WHERE user_id = $1`, userID)
	return err
}

// SaveToken сохраняет новый токен для пользователя.
func (r *TokenRepository) SaveToken(ctx context.Context, userID int, accessToken, refreshToken, userAgent, ipAddress string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tokens (user_id, access_token, refresh_token, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, accessToken, refreshToken, userAgent, ipAddress, expiresAt)
	return err
}

// RevokeToken удаляет токен по значению access_token.
func (r *TokenRepository) RevokeToken(ctx context.Context, accessToken string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tokens WHERE access_token = $1`, accessToken)
	return err
}

// IsTokenValid проверяет, существует ли access-токен в БД.
func (r *TokenRepository) IsTokenValid(ctx context.Context, accessToken string) (bool, error) {
	var expiresAt *time.Time
	err := r.db.QueryRowContext(ctx, `
		SELECT expires_at 
		FROM tokens 
		WHERE access_token = $1`, accessToken).Scan(&expiresAt)
	if err != nil {
		return false, fmt.Errorf("ошибка провкерки токена: %w", err)
	}
	// Проверяем, не истёк ли токен
	if expiresAt.Before(time.Now()) {
		return false, nil
	}
	return true, nil
}

// IsRefreshTokenValid проверяет, существует ли refresh-токен в БД.
func (r *TokenRepository) IsRefreshTokenValid(ctx context.Context, refreshToken string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 
			FROM tokens 
			WHERE refresh_token = $1
			)`, refreshToken).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("ошибка проверки токена: %w", err)
	}
	return exists, nil
}
