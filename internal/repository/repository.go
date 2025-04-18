package repository

import (
	domainChat "EduSync/internal/domain/chat"
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

// TeacherInitialsRepository описывает работу с таблицей teacher_initials.
type TeacherInitialsRepository interface {
	Upsert(ctx context.Context, initials string, teacherID *int, institutionID int) (int, error)
	GetByInitials(ctx context.Context, initials string, institutionID int) (*domainSchedule.TeacherInitials, error)
	GetByID(ctx context.Context, id int) (*domainSchedule.TeacherInitials, error)
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

// ChatRepository описывает операции для работы с чатами.
type ChatRepository interface {
	CreateChat(ctx context.Context, chat *domainChat.Chat) (int, error)
	GetChatByID(ctx context.Context, chatID int) (*domainChat.Chat, error)
	DeleteChat(ctx context.Context, chatID int) error
	UpdateChatInvite(ctx context.Context, chatID int, newJoinCode, newInviteLink string) error
	// GetParticipants Метод для получения списка участников чата.
	GetParticipants(ctx context.Context, chatID int) ([]*domainChat.Participant, error)
	RemoveParticipant(ctx context.Context, chatID int, userID int) error
	JoinChat(ctx context.Context, chatID, userID int) error
	LeaveChat(ctx context.Context, chatID int, userID int) error
	IsParticipant(ctx context.Context, chatID int, userID int) (bool, error)
	IsOwner(ctx context.Context, chatID int, userID int) (bool, error)
}

type MessageRepository interface {
	GetMessageByID(ctx context.Context, msgID int) (*domainChat.Message, error)
	CreateMessage(ctx context.Context, msg *domainChat.Message) (int, error)
	GetMessages(ctx context.Context, chatID int, limit, offset int) ([]*domainChat.Message, error)
	DeleteMessage(ctx context.Context, messageID int) error
	SearchMessages(ctx context.Context, chatID int, query string, limit, offset int) ([]*domainChat.Message, error)
	GetMessageFileInfo(ctx context.Context, messageID int) ([]*domainChat.FileInfo, error)

	BeginTx(ctx context.Context) (*sql.Tx, error)
	CreateMessageTx(ctx context.Context, tx *sql.Tx, msg *domainChat.Message) (int, error)
	CreateMessageFileTx(ctx context.Context, tx *sql.Tx, messageID int, fileURL string) error
	DeleteMessageFilesTx(ctx context.Context, tx *sql.Tx, messageID int) error
	DeleteMessageTx(ctx context.Context, tx *sql.Tx, messageID int) error
}
