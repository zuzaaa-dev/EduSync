package group

import (
	domainGroup "EduSync/internal/domain/group"
	"EduSync/internal/integration/parser/rksi"
	"EduSync/internal/repository"
	"fmt"
	"time"
)

type GroupService struct {
	repo          repository.GroupRepository
	parser        rksi.Parser // Интерфейс, описывающий методы парсинга групп
	institutionID int         // ID учебного заведения (для rksi)
}

// NewGroupService создает новый сервис групп.
func NewGroupService(repo repository.GroupRepository, parser rksi.Parser, institutionID int) *GroupService {
	return &GroupService{
		repo:          repo,
		parser:        parser,
		institutionID: institutionID,
	}
}

// UpdateGroups получает группы через парсер и сохраняет их в БД.
func (s *GroupService) UpdateGroups() error {
	// Парсим группы. Парсер внутри адаптера может возвращать слайс строк (названий групп)
	groupNames, err := s.parser.FetchGroups()
	if err != nil {
		return fmt.Errorf("ошибка парсинга групп: %v", err)
	}

	var groups []*domainGroup.Group
	for _, name := range groupNames {
		groups = append(groups, &domainGroup.Group{
			Name:          name,
			InstitutionID: s.institutionID,
		})
	}

	// Сохраняем группы в БД
	if err := s.repo.SaveGroups(groups); err != nil {
		return fmt.Errorf("ошибка сохранения групп: %v", err)
	}

	return nil
}

// GetGroupsByInstitutionID возвращает группы для заданного учреждения.
func (s *GroupService) GetGroupsByInstitutionID(institutionID int) ([]*domainGroup.Group, error) {
	return s.repo.GetByInstitutionID(institutionID)
}

// Запуск воркера для периодического обновления групп (например, раз в 24 часа).
func (s *GroupService) StartWorker(interval time.Duration) {
	go func() {
		for {
			err := s.UpdateGroups()
			if err != nil {
				fmt.Printf("Ошибка обновления групп: %v\n", err)
			} else {
				fmt.Println("Группы успешно обновлены")
			}
			time.Sleep(interval)
		}
	}()
}
