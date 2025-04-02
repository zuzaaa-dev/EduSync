package institution

import (
	domainInstitution "EduSync/internal/domain/institution"
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type emailMaskService struct {
	repo repository.EmailMaskRepository
	log  *logrus.Logger
}

func NewEmailMaskService(repo repository.EmailMaskRepository, logger *logrus.Logger) service.EmailMaskService {
	return &emailMaskService{
		repo: repo,
		log:  logger,
	}
}

func (s *emailMaskService) AllMasks(ctx context.Context) ([]*domainInstitution.EmailMask, error) {
	s.log.Info("Получение всех почтовых масок")
	masks, err := s.repo.All(ctx)
	if err != nil {
		s.log.Errorf("Ошибка получения почтовых масок: %v", err)
		return nil, fmt.Errorf("не удалось получить почтовые маски")
	}
	return masks, nil
}

func (s *emailMaskService) MaskByValue(ctx context.Context, mask string) (*domainInstitution.EmailMask, error) {
	s.log.Infof("Получение почтовой маски для: %s", mask)
	result, err := s.repo.ByEmailMask(ctx, mask)
	if err != nil {
		s.log.Errorf("Ошибка получения маски по значению %s: %v", mask, err)
		return nil, fmt.Errorf("не удалось получить маску")
	}
	if result == nil {
		return nil, fmt.Errorf("маска %s не найдена", mask)
	}
	return result, nil
}
