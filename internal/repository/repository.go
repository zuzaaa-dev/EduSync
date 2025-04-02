package repository

import (
	domainGroup "EduSync/internal/domain/group"
	domainInstitution "EduSync/internal/domain/institution"
	domainSchedule "EduSync/internal/domain/schedule"
	domainSubject "EduSync/internal/domain/subject"
	domainUser "EduSync/internal/domain/user"
	"context"
	"database/sql"
	"time"
)

// UserRepository описывает контракт для работы с пользователями.
type UserRepository interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	Create(ctx context.Context, tx *sql.Tx, user *domainUser.User) (int, error)
	ByEmail(ctx context.Context, email string) (*domainUser.User, error)
}

// StudentRepository описывает контракт для работы со студентами.
type StudentRepository interface {
	Create(ctx context.Context, tx *sql.Tx, userID, institutionID, groupID int) error
	ByUserID(ctx context.Context, userID int) (*domainUser.Student, error)
	ByGroupID(ctx context.Context, groupID int) ([]*domainUser.Student, error)
}

// TeacherRepository описывает контракт для работы с преподавателями.
type TeacherRepository interface {
	Create(ctx context.Context, tx *sql.Tx, userID, institutionID int) error
	ByUserID(ctx context.Context, userID int) (*domainUser.Teacher, error)
	ByInstitutionID(ctx context.Context, institutionID int) ([]*domainUser.Teacher, error)
	BySurname(ctx context.Context, surname string) ([]*domainUser.User, error)
}

type TokenRepository interface {
	DeleteForUser(ctx context.Context, userID int) error
	Save(ctx context.Context, userID int, accessToken, refreshToken, userAgent, ipAddress string, expiresAt time.Time) error
	Revoke(ctx context.Context, accessToken string) error
	IsValid(ctx context.Context, accessToken string) (bool, error)
	IsRefreshValid(ctx context.Context, refreshToken string) (bool, error)
}

// GroupRepository описывает контракт для работы с группами.
type GroupRepository interface {
	Save(ctx context.Context, groups []*domainGroup.Group) error
	ByInstitutionID(ctx context.Context, institutionID int) ([]*domainGroup.Group, error)
	ById(ctx context.Context, groupId int) (*domainGroup.Group, error)
	ByName(ctx context.Context, name string) (*domainGroup.Group, error)
}

// InstitutionRepository описывает контракт доступа к данным учебных заведений.
type InstitutionRepository interface {
	ByID(ctx context.Context, id int) (*domainInstitution.Institution, error)
	All(ctx context.Context) ([]*domainInstitution.Institution, error)
}

// EmailMaskRepository описывает контракт доступа к почтовым маскам.
type EmailMaskRepository interface {
	All(ctx context.Context) ([]*domainInstitution.EmailMask, error)
	ByEmailMask(ctx context.Context, mask string) (*domainInstitution.EmailMask, error)
}

type SubjectRepository interface {
	Create(ctx context.Context, name string, institutionID int) (int, error)
	ByID(ctx context.Context, id int) (*domainSubject.Subject, error)
	ByInstitutionID(ctx context.Context, institutionID int) ([]*domainSubject.Subject, error)
	ByGroupID(ctx context.Context, groupID int) ([]*domainSubject.Subject, error)
	ByNameAndInstitution(ctx context.Context, discipline string, id int) (*domainSubject.Subject, error)
}

// ScheduleRepository описывает контракт доступа к данным расписания.
type ScheduleRepository interface {
	Save(ctx context.Context, entries []*domainSchedule.Schedule) error
	ByGroupID(ctx context.Context, groupID int) ([]*domainSchedule.Schedule, error)
}
