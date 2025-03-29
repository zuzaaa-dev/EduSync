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

// CreateSubject создаёт предмет, если его ещё нет.
func (s *subjectService) CreateSubject(ctx context.Context, name string, institutionID int) (int, error) {
	s.log.Infof("Создание группы: %v, institutionID=%v", name, institutionID)
	if name == "" {
		return 0, errors.New("название предмета не может быть пустым")
	}
	return s.subjectRepo.Create(ctx, name, institutionID)
}

// GetSubjectByID получает предмет по ID.
func (s *subjectService) GetSubjectByID(ctx context.Context, id int) (*subject.Subject, error) {
	return s.subjectRepo.GetByID(ctx, id)
}

// GetSubjectsByInstitutionID получает список предметов по ID учебного заведения.
func (s *subjectService) GetSubjectsByInstitutionID(ctx context.Context, institutionID int) ([]*subject.Subject, error) {
	return s.subjectRepo.GetByInstitutionID(ctx, institutionID)
}

// GetSubjectsByGroupID получает список предметов по ID группы.
func (s *subjectService) GetSubjectsByGroupID(ctx context.Context, groupID int) ([]*subject.Subject, error) {
	return s.subjectRepo.GetByGroupID(ctx, groupID)
}

func (s *subjectService) GetSubjectByNameAndInstitution(ctx context.Context, discipline string, id int) (*subject.Subject, error) {
	return s.subjectRepo.GetByNameAndInstitution(ctx, discipline, id)
}
