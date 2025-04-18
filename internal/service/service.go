package service

import (
	domainChat "EduSync/internal/domain/chat"
	domainGroup "EduSync/internal/domain/group"
	domainInstitution "EduSync/internal/domain/institution"
	deliverSchedule "EduSync/internal/domain/schedule"
	domainSubject "EduSync/internal/domain/subject"
	domainUser "EduSync/internal/domain/user"
	"context"
	"github.com/gin-gonic/gin"
	"mime/multipart"
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
	ByInstitutionID(ctx context.Context, institutionID int) ([]*domainGroup.Group, error)
	ById(ctx context.Context, groupId int) (*domainGroup.Group, error)
	StartWorker(interval time.Duration)
	Update(ctx context.Context) error
}

// InstitutionService описывает методы работы с учебными заведениями.
type InstitutionService interface {
	ByID(ctx context.Context, id int) (*domainInstitution.Institution, error)
	All(ctx context.Context) ([]*domainInstitution.Institution, error)
}

// EmailMaskService описывает бизнес-логику для работы с почтовыми масками.
type EmailMaskService interface {
	AllMasks(ctx context.Context) ([]*domainInstitution.EmailMask, error)
	MaskByValue(ctx context.Context, mask string) (*domainInstitution.EmailMask, error)
}

type SubjectService interface {
	Create(ctx context.Context, name string, institutionID int) (int, error)
	ByID(ctx context.Context, id int) (*domainSubject.Subject, error)
	ByInstitutionID(ctx context.Context, institutionID int) ([]*domainSubject.Subject, error)
	ByGroupID(ctx context.Context, groupID int) ([]*domainSubject.Subject, error)
	ByNameAndInstitution(ctx context.Context, discipline string, id int) (*domainSubject.Subject, error)
}

// ScheduleService описывает методы работы с расписанием.
type ScheduleService interface {
	Update(ctx context.Context, groupName string) error
	ByGroupID(ctx context.Context, groupID int) ([]*deliverSchedule.Item, error)
	StartWorker(interval time.Duration)
	StartWorkerInitials(interval time.Duration)
}

type ChatService interface {
	CreateChat(ctx context.Context, c domainChat.Chat) (int, error)
	GetChatParticipants(ctx context.Context, chatID int) ([]*domainChat.Participant, error)
	JoinChat(ctx context.Context, chatID, userID int) error
	RecreateInvite(ctx context.Context, chatID int, ownerID int) error
	DeleteChat(ctx context.Context, chatID int, ownerID int) error
	RemoveParticipant(ctx context.Context, chatID int, ownerID int, participantID int) error
	LeaveChat(ctx context.Context, chatID int, userID int) error
}

type MessageService interface {
	GetMessages(ctx context.Context, chatID, limit, offset int) ([]*domainChat.Message, error)
	SendMessage(ctx context.Context, msg domainChat.Message) (int, error)
	DeleteMessage(ctx context.Context, messageID int, requesterID int) error
	ReplyMessage(ctx context.Context, parentMessageID int, msg domainChat.Message) (int, error)
	SearchMessages(ctx context.Context, chatID int, query string, limit, offset int) ([]*domainChat.Message, error)
	GetMessageFiles(ctx context.Context, messageID int) ([]*domainChat.FileInfo, error)
	SendMessageWithFiles(ctx context.Context, msg domainChat.Message, files []*multipart.FileHeader, c *gin.Context) (int, error)
}
