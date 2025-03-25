package service

import (
	domainGroup "EduSync/internal/domain/group"
	domainUser "EduSync/internal/domain/user"
)

type UserService interface {
	Register(user domainUser.CreateUser) (int, error)
	Login(email, password, userAgent, ipAddress string) (accessToken, refreshToken string, err error)
	Logout(accessToken string) error
	RefreshToken(inputRefreshToken, userAgent, ipAddress string) (accessToken string, refreshToken string, err error)
}

// GroupService описывает контракт для работы с группами.
type GroupService interface {
	UpdateGroups() error
	GetGroupsByInstitutionID(institutionID int) ([]*domainGroup.Group, error)
	GetGroupById(groupId int) (*domainGroup.Group, error)
}
