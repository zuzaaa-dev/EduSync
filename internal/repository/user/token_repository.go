package user

import (
	"database/sql"
	"time"
)

// TokenRepository обеспечивает работу с таблицей токенов.
type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// DeleteTokensForUser удаляет все токены для пользователя.
// Это позволит гарантировать, что при новом логине не будет дублирования.
func (r *TokenRepository) DeleteTokensForUser(userID int) error {
	_, err := r.db.Exec(`DELETE FROM tokens WHERE user_id = $1`, userID)
	return err
}

// SaveToken сохраняет новый токен для пользователя.
func (r *TokenRepository) SaveToken(userID int, accessToken, refreshToken, userAgent, ipAddress string, expiresAt time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO tokens (user_id, access_token, refresh_token, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, accessToken, refreshToken, userAgent, ipAddress, expiresAt)
	return err
}

// RevokeToken удаляет токен по значению access_token.
func (r *TokenRepository) RevokeToken(accessToken string) error {
	_, err := r.db.Exec(`DELETE FROM tokens WHERE access_token = $1`, accessToken)
	return err
}

// IsTokenValid проверяет, существует ли access-токен в БД.
func (r *TokenRepository) IsTokenValid(accessToken string) (bool, error) {
	var expiresAt time.Time
	err := r.db.QueryRow(`SELECT expires_at FROM tokens WHERE access_token = $1`, accessToken).Scan(&expiresAt)
	if err != nil {
		return false, err
	}
	// Проверяем, не истёк ли токен
	if expiresAt.Before(time.Now()) {
		return false, nil
	}
	return true, nil
}

// IsRefreshTokenValid проверяет, существует ли refresh-токен в БД.
func (r *TokenRepository) IsRefreshTokenValid(refreshToken string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM tokens WHERE refresh_token = $1)`, refreshToken).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
