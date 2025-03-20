package user

import "time"
import "github.com/golang-jwt/jwt/v5"

// TokenClaims — данные, которые будут помещаться в JWT.
type TokenClaims struct {
	ID        int    `json:"id"`
	IsTeacher bool   `json:"is_teacher"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	jwt.RegisteredClaims
}

type TokenRepository interface {
	DeleteTokensForUser(userID int) error
	SaveToken(userID int, accessToken, refreshToken, userAgent, ipAddress string, expiresAt time.Time) error
	RevokeToken(accessToken string) error
	IsTokenValid(accessToken string) (bool, error)
	IsRefreshTokenValid(refreshToken string) (bool, error)
}
