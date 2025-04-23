package material

import (
	"EduSync/internal/service"
	"context"
	"embed"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"

	"EduSync/internal/repository"
)

type fileService struct {
	files repository.FileRepository
	msgs  repository.MessageRepository
	chats repository.ChatRepository
	log   *logrus.Logger
}

func NewFileService(
	files repository.FileRepository,
	msgs repository.MessageRepository,
	chats repository.ChatRepository,
	log *logrus.Logger,
) service.FileService {
	return &fileService{files, msgs, chats, log}
}

var f embed.FS

func (s *fileService) File(ctx context.Context, userID, fileID int) (*os.File, string, error) {
	f, err := s.files.ByID(ctx, fileID)
	if err != nil {
		s.log.Errorf("fileService.File.ByID: %v", err)
		return nil, "", fmt.Errorf("internal error")
	}
	if f == nil {
		return nil, "", fmt.Errorf("file not found")
	}

	msg, err := s.msgs.ByID(ctx, f.MessageID)
	if err != nil {
		s.log.Errorf("fileService.File.GetMsg: %v", err)
		return nil, "", fmt.Errorf("internal error")
	}
	if msg == nil {
		return nil, "", fmt.Errorf("message not found")
	}

	ok, err := s.chats.IsParticipant(ctx, msg.ChatID, userID)
	if err != nil {
		s.log.Errorf("fileService.File.IsParticipant: %v", err)
		return nil, "", fmt.Errorf("internal error")
	}
	if !ok {
		return nil, "", fmt.Errorf("permission denied")
	}

	fhandle, err := os.Open(f.FileURL)
	if err != nil {
		s.log.Errorf("fileService.File.Open: %v", err)
		return nil, "", fmt.Errorf("internal error")
	}

	return fhandle, fhandle.Name(), nil
	//panic("implement me")
}
