package schedule

import (
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"

	domainSch "EduSync/internal/domain/schedule"
)

type teacherInitialsService struct {
	repo repository.TeacherInitialsRepository
	log  *logrus.Logger
}

func NewTeacherInitialsService(repo repository.TeacherInitialsRepository, log *logrus.Logger) service.TeacherInitialsService {
	return &teacherInitialsService{repo: repo, log: log}
}

func (s *teacherInitialsService) List(ctx context.Context, institutionID int) ([]*domainSch.TeacherInitials, error) {
	list, err := s.repo.GetAll(ctx, institutionID)
	if err != nil {
		return nil, fmt.Errorf("service.List teacher_initials: %w", err)
	}
	return list, nil
}
