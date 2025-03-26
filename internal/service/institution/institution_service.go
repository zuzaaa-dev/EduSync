package institution

import (
	domainInstitution "EduSync/internal/domain/institution"
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"

	"github.com/sirupsen/logrus"
)

type Service struct {
	repo repository.InstitutionRepository
	log  *logrus.Logger
}

func NewInstitutionService(repo repository.InstitutionRepository, logger *logrus.Logger) service.InstitutionService {
	return &Service{
		repo: repo,
		log:  logger,
	}
}

func (s *Service) GetInstitutionByID(ctx context.Context, id int) (*domainInstitution.Institution, error) {
	s.log.Infof("Получение учреждения по id: %d", id)
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetAllInstitutions(ctx context.Context) ([]*domainInstitution.Institution, error) {
	s.log.Info("Получение списка всех учебных заведений")
	return s.repo.GetAll(ctx)
}
