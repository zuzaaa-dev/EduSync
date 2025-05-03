package service

import (
	dtoChat "EduSync/internal/delivery/dto/chat"
	dtoChat2 "EduSync/internal/delivery/http/chat/dto"
	dtoFavorite "EduSync/internal/delivery/http/favorite/dto"
	dtoSchedule "EduSync/internal/delivery/http/schedule/dto"
	domainChat "EduSync/internal/domain/chat"
	domainGroup "EduSync/internal/domain/group"
	domainInstitution "EduSync/internal/domain/institution"
	deliverSchedule "EduSync/internal/domain/schedule"
	domainSchedule "EduSync/internal/domain/schedule"
	domainSubject "EduSync/internal/domain/subject"
	domainUser "EduSync/internal/domain/user"
	"context"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"os"
	"time"
)

type UserService interface {
	Register(ctx context.Context, user domainUser.CreateUser) (int, error)
	Login(ctx context.Context, email, password, userAgent, ipAddress string) (accessToken, refreshToken string, err error)
	UpdateProfile(ctx context.Context, u domainUser.UpdateUser, userAgent string, ipAddress string) (string, string, error)
	Logout(ctx context.Context, accessToken string) error
	RefreshToken(ctx context.Context, inputRefreshToken, userAgent, ipAddress string) (accessToken, refreshToken string, err error)
	FindTeacherByName(ctx context.Context, teacher string) (*domainUser.User, error)
}

type TeacherInitialsService interface {
	List(ctx context.Context, institutionID int) ([]*domainSchedule.TeacherInitials, error)
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
	Save(ctx context.Context, groupName string) error
	Create(ctx context.Context, req *dtoSchedule.CreateScheduleReq) (int, error)
	ByGroupID(ctx context.Context, groupID int) ([]*deliverSchedule.Item, error)
	ByTeacherInitialsID(ctx context.Context, initialsID int) ([]*deliverSchedule.Item, error)
	ByID(ctx context.Context, id int) (*domainSchedule.Schedule, error)
	Update(ctx context.Context, id int, req *dtoSchedule.UpdateScheduleReq) error
	Delete(ctx context.Context, id int) error
	StartWorker(interval time.Duration)
	StartWorkerInitials(interval time.Duration)
}

type ChatService interface {
	CreateChat(ctx context.Context, c domainChat.Chat) (*domainChat.Chat, error)
	ListForUser(ctx context.Context, userID int, isTeacher bool) ([]*dtoChat.ChatInfo, error)
	RecreateInvite(ctx context.Context, chatID int, ownerID int) (*domainChat.Chat, error)
	ChatParticipants(ctx context.Context, chatID int) ([]*domainChat.Participant, error)
	JoinChat(ctx context.Context, userID int, code string) (*dtoChat.ChatInfo, error)
	DeleteChat(ctx context.Context, chatID int, ownerID int) error
	RemoveParticipant(ctx context.Context, chatID int, ownerID int, participantID int) error
	LeaveChat(ctx context.Context, chatID int, userID int) error
}

type MessageService interface {
	Messages(ctx context.Context, chatID, limit, offset int) ([]*domainChat.Message, error)
	SendMessage(ctx context.Context, msg domainChat.Message) (int, error)
	DeleteMessage(ctx context.Context, messageID int, requesterID int) error
	ReplyMessage(ctx context.Context, parentMessageID int, msg domainChat.Message) (int, error)
	SearchMessages(ctx context.Context, chatID int, query string, limit, offset int) ([]*domainChat.Message, error)
	MessageFiles(ctx context.Context, messageID int) ([]*domainChat.FileInfo, error)
	SendMessageWithFiles(ctx context.Context, msg domainChat.Message, files []*multipart.FileHeader, c *gin.Context) (int, error)
}

// FileService отдаёт файл по id, проверяя, что пользователь — участник чата.
type FileService interface {
	File(ctx context.Context, userID, fileID int) (*os.File, string, error)
}

type FileFavoriteService interface {
	AddFavorite(ctx context.Context, userID, fileID int) error
	RemoveFavorite(ctx context.Context, userID, fileID int) error
	ListFavorites(ctx context.Context, userID int) ([]dtoFavorite.FileInfo, error)
}

type PollService interface {
	CreatePoll(ctx context.Context, userID, chatID int, question string, options []string) (int, error)
	DeletePoll(ctx context.Context, userID, pollID int) error
	Vote(ctx context.Context, userID, pollID, optionID int) error
	ListPolls(ctx context.Context, userID, chatID, limit, offset int) ([]*dtoChat2.PollSummary, error)
	Unvote(ctx context.Context, userID, pollID, optionID int) error
}
