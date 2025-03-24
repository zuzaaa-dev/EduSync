package repository

import (
	domainGroup "EduSync/internal/domain/group"
	domainUser "EduSync/internal/domain/user"
	"time"
)

// UserRepository описывает контракт для работы с пользователями.
type UserRepository interface {
	CreateUser(user *domainUser.User) (int, error)
	GetUserByEmail(email string) (*domainUser.User, error)
}

type TokenRepository interface {
	DeleteTokensForUser(userID int) error
	SaveToken(userID int, accessToken, refreshToken, userAgent, ipAddress string, expiresAt time.Time) error
	RevokeToken(accessToken string) error
	IsTokenValid(accessToken string) (bool, error)
	IsRefreshTokenValid(refreshToken string) (bool, error)
}

// GroupRepository описывает контракт для работы с группами.
type GroupRepository interface {
	SaveGroups(groups []*domainGroup.Group) error
	GetByInstitutionID(institutionID int) ([]*domainGroup.Group, error)
}
