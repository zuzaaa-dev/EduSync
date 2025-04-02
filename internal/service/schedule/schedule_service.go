package schedule

import (
	"EduSync/internal/repository"
	"context"
	"fmt"
	"time"

	domainSchedule "EduSync/internal/domain/schedule"
	"EduSync/internal/integration/parser/rksi/schedule"
	"EduSync/internal/service"
	"github.com/sirupsen/logrus"
)

type scheduleService struct {
	repo       repository.ScheduleRepository
	parser     *schedule.ScheduleParser // Интерфейс адаптера парсинга расписания
	subjectSvc service.SubjectService   // Сервис для работы с предметами
	userSvc    service.UserService      // Сервис для работы с пользователями (для поиска преподавателя)
	groupRepo  repository.GroupRepository
	log        *logrus.Logger
}

// NewScheduleService создает новый сервис для расписания.
func NewScheduleService(
	repo repository.ScheduleRepository,
	parser *schedule.ScheduleParser,
	subjectSvc service.SubjectService,
	userSvc service.UserService,
	groupRepo repository.GroupRepository,
	log *logrus.Logger,
) service.ScheduleService {
	return &scheduleService{
		repo:       repo,
		parser:     parser,
		subjectSvc: subjectSvc,
		userSvc:    userSvc,
		groupRepo:  groupRepo,
		log:        log,
	}
}

// Update получает расписание с сайта, выполняет проверки и сохраняет его в БД.
func (s *scheduleService) Update(ctx context.Context, groupName string) error {
	s.log.Infof("Обновление расписания для группы: %s (id: %d)", groupName)

	// Получаем расписание с помощью парсера
	parsedEntries, err := s.parser.FetchSchedule(ctx, groupName)
	if err != nil {
		s.log.Errorf("ошибка парсинга расписания: %v", err)
		return fmt.Errorf("ошибка парсинга расписания")
	}

	group, err := s.groupRepo.ByName(ctx, groupName)
	if err != nil {
		s.log.Errorf("ошибка получения id группы: %v", err)
		return fmt.Errorf("ошибка получения группы")
	}
	groupID := group.ID
	dateCounters := make(map[time.Time]int, 7)
	var entries []*domainSchedule.Schedule
	for _, pe := range parsedEntries {
		// Если для этой даты еще нет счетчика, начинаем с 1
		dateCounters[pe.Date]++
		pairNumber := dateCounters[pe.Date]
		// Преобразование даты и времени из строк в time.Time
		// Предполагаем, что формат даты "02.01.2006" и времени "15:04"
		startTime, err := time.Parse("15:04", pe.StartTime)
		if err != nil {
			s.log.Errorf("ошибка парсинга startTime (%s): %v", pe.StartTime, err)
			continue
		}
		endTime, err := time.Parse("15:04", pe.EndTime)
		if err != nil {
			s.log.Errorf("ошибка парсинга endTime (%s): %v", pe.EndTime, err)
			continue
		}

		// Проверяем предмет: если не существует, добавляем его через subjectSvc
		subj, err := s.subjectSvc.ByNameAndInstitution(ctx, pe.Discipline, s.parser.InstitutionID)
		if err != nil {
			s.log.Errorf("ошибка поиска предмета: %v", err)
			continue
		}
		if subj == nil {
			subjID, err := s.subjectSvc.Create(ctx, pe.Discipline, s.parser.InstitutionID)
			if err != nil {
				s.log.Errorf("ошибка создания предмета: %v", err)
				continue
			}
			// Получаем предмет после создания
			subj, err = s.subjectSvc.ByID(ctx, subjID)
			if err != nil {
				s.log.Errorf("ошибка получения предмета: %v", err)
				continue
			}
		}

		// Проверяем преподавателя: если в расписании указано ФИО, ищем преподавателя по ФИО.
		// Здесь можно добавить собственную логику поиска (например, через userSvc)
		var teacherID *int
		teacherInitials := pe.Teacher
		// Попытаемся найти преподавателя, если имя не пустое и не равно "-"
		if pe.Teacher != "" && pe.Teacher != "-" {
			t, err := s.userSvc.FindTeacherByName(ctx, pe.Teacher)
			if err != nil {
				s.log.Errorf("ошибка поиска преподавателя: %v", err)
			} else if t != nil {
				teacherID = &t.ID
				teacherInitials = "" // Если преподаватель найден, можно не сохранять инициалы
			}
		}

		entry := &domainSchedule.Schedule{
			GroupID:         groupID,
			SubjectID:       subj.ID,
			Date:            pe.Date,
			PairNumber:      pairNumber, // Здесь можно добавить парсинг номера пары, если он есть
			Classroom:       pe.Classroom,
			TeacherID:       teacherID,
			TeacherInitials: teacherInitials,
			StartTime:       combineTime(pe.Date, startTime),
			EndTime:         combineTime(pe.Date, endTime),
		}

		entries = append(entries, entry)
	}

	// Сохраняем расписание в БД
	err = s.repo.Save(ctx, entries)
	if err != nil {
		s.log.Errorf("ошибка сохранения расписания: %v", err)
		return fmt.Errorf("ошибка сохранения расписания")
	}

	return nil
}

// combineTime объединяет дату и время в один объект time.Time.
func combineTime(date, t time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), t.Second(), 0, date.Location())
}

// ByGroupID возвращает расписание для заданной группы.
func (s *scheduleService) ByGroupID(ctx context.Context, groupID int) ([]*domainSchedule.Schedule, error) {
	return s.repo.ByGroupID(ctx, groupID)
}

// StartWorker запускает периодическое обновление расписания для заданной группы.
func (s *scheduleService) StartWorker(interval time.Duration, groupID int, groupName string) {
	go func() {
		for {
			err := s.Update(context.Background(), groupName)
			if err != nil {
				s.log.Errorf("Ошибка обновления расписания: %v", err)
			} else {
				s.log.Info("Расписание успешно обновлено")
			}
			time.Sleep(interval)
		}
	}()
}
