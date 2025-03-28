package service

import (
	domainGroup "EduSync/internal/domain/group"
	domainInstitution "EduSync/internal/domain/institution"
	domainUser "EduSync/internal/domain/user"
	"context"
	"time"
)

type UserService interface {
	Register(ctx context.Context, user domainUser.CreateUser) (int, error)
	Login(ctx context.Context, email, password, userAgent, ipAddress string) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, accessToken string) error
	RefreshToken(ctx context.Context, inputRefreshToken, userAgent, ipAddress string) (accessToken string, refreshToken string, err error)
}

// GroupService описывает контракт для работы с группами.
type GroupService interface {
	GetGroupsByInstitutionID(ctx context.Context, institutionID int) ([]*domainGroup.Group, error)
	GetGroupById(ctx context.Context, groupId int) (*domainGroup.Group, error)
	StartWorker(interval time.Duration)
	UpdateGroups() error
}

// InstitutionService описывает методы работы с учебными заведениями.
type InstitutionService interface {
	GetInstitutionByID(ctx context.Context, id int) (*domainInstitution.Institution, error)
	GetAllInstitutions(ctx context.Context) ([]*domainInstitution.Institution, error)
}
