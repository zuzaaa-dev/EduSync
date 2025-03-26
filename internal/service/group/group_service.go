package group

import (
	domainGroup "EduSync/internal/domain/group"
	"EduSync/internal/integration/parser/rksi"
	"EduSync/internal/repository"
	"context"
	"github.com/sirupsen/logrus"
	"time"
)

type Service struct {
	repo   repository.GroupRepository
	parser rksi.Parser // Интерфейс, описывающий методы парсинга групп
	log    *logrus.Logger
}

// NewGroupService создает новый сервис групп.
func NewGroupService(repo repository.GroupRepository, parser rksi.Parser, logger *logrus.Logger) *Service {
	return &Service{
		repo:   repo,
		parser: parser,
		log:    logger,
	}
}

// UpdateGroups получает группы через парсер и сохраняет их в БД.
func (s *Service) UpdateGroups() error {
	s.log.Info("Обновление групп")
	// Парсим группы. Парсер внутри адаптера может возвращать слайс строк (названий групп)
	groupNames, institutionId, err := s.parser.FetchGroups()
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
	if err := s.repo.SaveGroups(context.Background(), groups); err != nil {
		s.log.Errorf("Ошибка парсинга групп: %v", err)
		return err
	}

	return nil
}

// GetGroupsByInstitutionID возвращает группы для заданного учреждения.
func (s *Service) GetGroupsByInstitutionID(ctx context.Context, institutionID int) ([]*domainGroup.Group, error) {
	s.log.Infof("Получение группы по id учреждения: %d", institutionID)
	return s.repo.GetByInstitutionID(ctx, institutionID)
}

func (s *Service) GetGroupById(ctx context.Context, groupId int) (*domainGroup.Group, error) {
	s.log.Infof("Получение группы по id группы: %d", groupId)
	return s.repo.GetById(ctx, groupId)
}

// Запуск воркера для периодического обновления групп (например, раз в 24 часа).
func (s *Service) StartWorker(interval time.Duration) {
	go func() {
		for {
			err := s.UpdateGroups()
			if err != nil {
				s.log.Errorf("Ошибка обновления групп: %v\n", err)
			} else {
				s.log.Infof("Группы успешно обновлены")
			}
			time.Sleep(interval)
		}
	}()
}
