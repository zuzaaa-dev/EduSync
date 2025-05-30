package group

import (
	domainGroup "EduSync/internal/domain/group"
	"EduSync/internal/integration/parser/rksi/group"
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"github.com/sirupsen/logrus"
	"time"
)

type Service struct {
	repo   repository.GroupRepository
	parser group.Parser // Интерфейс, описывающий методы парсинга групп
	log    *logrus.Logger
}

// NewGroupService создает новый сервис групп.
func NewGroupService(repo repository.GroupRepository, parser group.Parser, logger *logrus.Logger) service.GroupService {
	return &Service{
		repo:   repo,
		parser: parser,
		log:    logger,
	}
}

// Update получает группы через парсер и сохраняет их в БД.
func (s *Service) Update(ctx context.Context) error {
	s.log.Info("Обновление групп")

	groupNames, institutionId, err := s.parser.FetchGroups(ctx)
	if err != nil {
		s.log.Errorf("Ошибка парсинга групп: %v", err)
		return err
	}

	var groups []*domainGroup.Group
	for _, name := range groupNames {
		groups = append(groups, &domainGroup.Group{
			Name:          name,
			InstitutionID: institutionId,
		})
	}

	// Сохраняем группы в БД
	if err := s.repo.Save(context.Background(), groups); err != nil {
		s.log.Errorf("Ошибка парсинга групп: %v", err)
		return err
	}

	return nil
}

// ByInstitutionID возвращает группы для заданного учреждения.
func (s *Service) ByInstitutionID(ctx context.Context, institutionID int) ([]*domainGroup.Group, error) {
	s.log.Infof("Получение группы по id учреждения: %d", institutionID)
	return s.repo.ByInstitutionID(ctx, institutionID)
}

func (s *Service) ById(ctx context.Context, groupId int) (*domainGroup.Group, error) {
	s.log.Infof("Получение группы по id группы: %d", groupId)
	return s.repo.ById(ctx, groupId)
}

// Запуск воркера для периодического обновления групп (например, раз в 24 часа).
func (s *Service) StartWorker(interval time.Duration) {
	ctx := context.Background()
	go func(ctx context.Context) {
		for {
			err := s.Update(ctx)
			if err != nil {
				s.log.Errorf("Ошибка обновления групп: %v\n", err)
			} else {
				s.log.Infof("Группы успешно обновлены")
			}
			time.Sleep(interval)
		}
	}(ctx)
}
