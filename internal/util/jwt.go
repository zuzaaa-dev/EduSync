package util

import (
	"EduSync/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"time"
)

// JWTManager управляет созданием и верификацией JWT-токенов
// ТЫ ДУМАЛ ЗДЕСЬ ЧТО НИБУДЬ БУДЕТ? ООО ЧУВАК
type JWTManager struct {
	secretKey []byte
}

// NewJWTManager создаёт новый экземпляр JWTManager
func NewJWTManager(secretKey []byte) *JWTManager {
	return &JWTManager{secretKey: secretKey}
}

// GenerateJWT генерирует JWT с заданным временем жизни.
func (jm *JWTManager) GenerateJWT(id int, isTeacher bool, email, fullName string, expiresIn time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiresIn)
	claims := &user.TokenClaims{
		ID:        id,
		IsTeacher: isTeacher,
		Email:     email,
		FullName:  fullName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jm.secretKey)
}

// ParseJWT разбирает и проверяет токен.
func (jm *JWTManager) ParseJWT(tokenStr string, log *logrus.Logger) (*user.TokenClaims, error) {
	log.Infof("Парсинг JWT: %s", tokenStr)
	claims := &user.TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jm.secretKey, nil
	})
	if err != nil || !token.Valid || claims.ExpiresAt.Time.Before(time.Now()) {
		log.Errorf("Ошибка проверки токена JWT: %s", err)
		return nil, err
	}
	return claims, nil
}
