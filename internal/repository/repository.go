package repository

import (
	domainGroup "EduSync/internal/domain/group"
	domainInstitution "EduSync/internal/domain/institution"
	domainUser "EduSync/internal/domain/user"
	"context"
	"time"
)

// UserRepository описывает контракт для работы с пользователями.
type UserRepository interface {
	CreateUser(ctx context.Context, user *domainUser.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (*domainUser.User, error)
}

type TokenRepository interface {
	DeleteTokensForUser(ctx context.Context, userID int) error
	SaveToken(ctx context.Context, userID int, accessToken, refreshToken, userAgent, ipAddress string, expiresAt time.Time) error
	RevokeToken(ctx context.Context, accessToken string) error
	IsTokenValid(ctx context.Context, accessToken string) (bool, error)
	IsRefreshTokenValid(ctx context.Context, refreshToken string) (bool, error)
}

// GroupRepository описывает контракт для работы с группами.
type GroupRepository interface {
	SaveGroups(ctx context.Context, groups []*domainGroup.Group) error
	GetByInstitutionID(ctx context.Context, institutionID int) ([]*domainGroup.Group, error)
	GetById(ctx context.Context, groupId int) (*domainGroup.Group, error)
}

// InstitutionRepository описывает контракт доступа к данным учебных заведений.
type InstitutionRepository interface {
	GetByID(ctx context.Context, id int) (*domainInstitution.Institution, error)
	GetAll(ctx context.Context) ([]*domainInstitution.Institution, error)
}
