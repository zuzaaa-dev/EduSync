package favorite

import (
	dtoFavorite "EduSync/internal/delivery/http/favorite/dto"
	"EduSync/internal/domain/chat"
	"EduSync/internal/repository"
	"EduSync/internal/repository/favorite"
	"EduSync/internal/service"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
)

type fileFavoriteService struct {
	favRepo  repository.FileFavoriteRepository
	fileRepo repository.FileRepository
	msgRepo  repository.MessageRepository
	chatRepo repository.ChatRepository
	logger   *logrus.Logger
}

func NewFileFavoriteService(
	fav repository.FileFavoriteRepository,
	fr repository.FileRepository,
	mr repository.MessageRepository,
	cr repository.ChatRepository,
	log *logrus.Logger,
) service.FileFavoriteService {
	return &fileFavoriteService{
		favRepo:  fav,
		fileRepo: fr,
		msgRepo:  mr,
		chatRepo: cr,
		logger:   log,
	}
}

func (s *fileFavoriteService) AddFavorite(ctx context.Context, userID, fileID int) error {
	f, err := s.fileRepo.ByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("AddFavorite: %w", err)
	}
	if f == nil {
		return chat.ErrNotFound
	}

	msg, err := s.msgRepo.ByID(ctx, f.MessageID)
	if err != nil {
		return fmt.Errorf("AddFavorite: %w", err)
	}
	ok, err := s.chatRepo.IsParticipant(ctx, msg.ChatID, userID)
	if err != nil {
		return fmt.Errorf("AddFavorite: %w", err)
	}
	if !ok {
		return chat.ErrPermissionDenied
	}

	if err := s.favRepo.Add(ctx, userID, fileID); err != nil {
		if err == favorite.ErrAlreadyFavorited {
			return chat.ErrAlreadyFavorited
		}
		return fmt.Errorf("AddFavorite: %w", err)
	}
	return nil
}

func (s *fileFavoriteService) RemoveFavorite(ctx context.Context, userID, fileID int) error {
	f, err := s.fileRepo.ByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("RemoveFavorite: %w", err)
	}
	if f == nil {
		return chat.ErrNotFound
	}
	msg, err := s.msgRepo.ByID(ctx, f.MessageID)
	if err != nil {
		return fmt.Errorf("RemoveFavorite: %w", err)
	}
	ok, err := s.chatRepo.IsParticipant(ctx, msg.ChatID, userID)
	if err != nil {
		return fmt.Errorf("RemoveFavorite: %w", err)
	}
	if !ok {
		return chat.ErrPermissionDenied
	}
	// Удаляем
	if err := s.favRepo.Remove(ctx, userID, fileID); err != nil {
		if err == favorite.ErrFavoriteNotFound {
			return chat.ErrNotFavorited
		}
		return fmt.Errorf("RemoveFavorite: %w", err)
	}
	return nil
}

func (s *fileFavoriteService) ListFavorites(ctx context.Context, userID int) ([]dtoFavorite.FileInfo, error) {
	s.logger.Infof("ListFavorites: user=%d", userID)

	fileIDs, err := s.favRepo.ListByUser(ctx, userID)
	if err != nil {
		s.logger.Errorf("ListFavorites: ListByUser error: %v", err)
		return nil, fmt.Errorf("не удалось получить избранные файлы")
	}

	var out []dtoFavorite.FileInfo
	for _, fid := range fileIDs {
		f, err := s.fileRepo.ByID(ctx, fid)
		if err != nil {
			s.logger.Errorf("ListFavorites: FileRepo.ByID(%d): %v", fid, err)
			continue
		}
		if f == nil {
			// может удалён
			continue
		}
		out = append(out, dtoFavorite.FileInfo{ID: f.ID, FileURL: f.FileURL})
	}
	return out, nil
}
