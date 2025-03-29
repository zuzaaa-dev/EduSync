package service

import (
	domainGroup "EduSync/internal/domain/group"
	domainInstitution "EduSync/internal/domain/institution"
	domainSchedule "EduSync/internal/domain/schedule"
	"EduSync/internal/domain/subject"
	domainUser "EduSync/internal/domain/user"
	"context"
	"time"
)

type UserService interface {
	Register(ctx context.Context, user domainUser.CreateUser) (int, error)
	Login(ctx context.Context, email, password, userAgent, ipAddress string) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, accessToken string) error
	RefreshToken(ctx context.Context, inputRefreshToken, userAgent, ipAddress string) (accessToken, refreshToken string, err error)
	FindTeacherByName(ctx context.Context, teacher string) (*domainUser.User, error)
}

// GroupService описывает контракт для работы с группами.
type GroupService interface {
	GetGroupsByInstitutionID(ctx context.Context, institutionID int) ([]*domainGroup.Group, error)
	GetGroupById(ctx context.Context, groupId int) (*domainGroup.Group, error)
	StartWorker(interval time.Duration)
	UpdateGroups(ctx context.Context) error
}

// InstitutionService описывает методы работы с учебными заведениями.
type InstitutionService interface {
	GetInstitutionByID(ctx context.Context, id int) (*domainInstitution.Institution, error)
	GetAllInstitutions(ctx context.Context) ([]*domainInstitution.Institution, error)
}

type SubjectService interface {
	CreateSubject(ctx context.Context, name string, institutionID int) (int, error)
	GetSubjectByID(ctx context.Context, id int) (*subject.Subject, error)
	GetSubjectsByInstitutionID(ctx context.Context, institutionID int) ([]*subject.Subject, error)
	GetSubjectsByGroupID(ctx context.Context, groupID int) ([]*subject.Subject, error)
	GetSubjectByNameAndInstitution(ctx context.Context, discipline string, id int) (*subject.Subject, error)
}

// ScheduleService описывает методы работы с расписанием.
type ScheduleService interface {
	UpdateSchedule(ctx context.Context, groupName string) error
	GetScheduleByGroupID(ctx context.Context, groupID int) ([]*domainSchedule.Schedule, error)
	StartWorker(interval time.Duration, groupID int, groupName string)
}
