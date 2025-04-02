package subject

import (
	"EduSync/internal/domain/subject"
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
)

// subjectService предоставляет бизнес-логику для предметов.
type subjectService struct {
	subjectRepo repository.SubjectRepository
	log         *logrus.Logger
}

// NewSubjectService создаёт новый экземпляр сервиса предметов.
func NewSubjectService(subjectRepo repository.SubjectRepository, log *logrus.Logger) service.SubjectService {
	return &subjectService{subjectRepo: subjectRepo, log: log}
}

// Create создаёт предмет, если его ещё нет.
func (s *subjectService) Create(ctx context.Context, name string, institutionID int) (int, error) {
	s.log.Infof("Создание группы: %v, institutionID=%v", name, institutionID)
	if name == "" {
		return 0, errors.New("название предмета не может быть пустым")
	}
	return s.subjectRepo.Create(ctx, name, institutionID)
}

// ByID получает предмет по ID.
func (s *subjectService) ByID(ctx context.Context, id int) (*subject.Subject, error) {
	return s.subjectRepo.ByID(ctx, id)
}

// ByInstitutionID получает список предметов по ID учебного заведения.
func (s *subjectService) ByInstitutionID(ctx context.Context, institutionID int) ([]*subject.Subject, error) {
	return s.subjectRepo.ByInstitutionID(ctx, institutionID)
}

// ByGroupID получает список предметов по ID группы.
func (s *subjectService) ByGroupID(ctx context.Context, groupID int) ([]*subject.Subject, error) {
	return s.subjectRepo.ByGroupID(ctx, groupID)
}

func (s *subjectService) ByNameAndInstitution(ctx context.Context, discipline string, id int) (*subject.Subject, error) {
	return s.subjectRepo.ByNameAndInstitution(ctx, discipline, id)
}
