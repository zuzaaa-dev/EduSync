package schedule

import (
	dtoSchedule "EduSync/internal/delivery/http/schedule/dto"
	domainSchedule "EduSync/internal/domain/schedule"
	"EduSync/internal/integration/parser/rksi/schedule"
	"EduSync/internal/integration/parser/rksi/teacher"
	"EduSync/internal/repository"
	"EduSync/internal/service"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type scheduleService struct {
	repo                repository.ScheduleRepository
	parserSchedule      *schedule.ScheduleParser
	parserTeacher       *teacher.TeacherParser
	subjectSvc          service.SubjectService
	userSvc             service.UserService
	groupRepo           repository.GroupRepository
	teacherInitialsRepo repository.TeacherInitialsRepository
	log                 *logrus.Logger
}

// NewScheduleService создает новый сервис для расписания.
func NewScheduleService(
	repo repository.ScheduleRepository,
	parserSchedule *schedule.ScheduleParser,
	parserTeacher *teacher.TeacherParser,
	subjectSvc service.SubjectService,
	userSvc service.UserService,
	groupRepo repository.GroupRepository,
	teacherInitialsRepo repository.TeacherInitialsRepository,
	log *logrus.Logger,
) service.ScheduleService {
	return &scheduleService{
		repo:                repo,
		parserSchedule:      parserSchedule,
		parserTeacher:       parserTeacher,
		subjectSvc:          subjectSvc,
		userSvc:             userSvc,
		groupRepo:           groupRepo,
		teacherInitialsRepo: teacherInitialsRepo,
		log:                 log,
	}
}

// Save получает расписание с сайта, выполняет проверки и сохраняет его в БД.
func (s *scheduleService) Save(ctx context.Context, groupName string) error {
	s.log.Infof("Обновление расписания для группы: %s (id: %d)", groupName)

	// Получаем расписание с помощью парсера
	parsedEntries, err := s.parserSchedule.FetchSchedule(ctx, groupName)
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
		subj, err := s.subjectSvc.ByNameAndInstitution(ctx, pe.Discipline, s.parserSchedule.InstitutionID)
		if err != nil {
			s.log.Errorf("ошибка поиска предмета: %v", err)
			continue
		}
		if subj == nil {
			subjID, err := s.subjectSvc.Create(ctx, pe.Discipline, s.parserSchedule.InstitutionID)
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
		var teacherID int
		teacherInitials := pe.Teacher
		// Попытаемся найти преподавателя, если имя не пустое и не равно "-"
		if pe.Teacher != "" && pe.Teacher != "-" {
			t, err := s.userSvc.FindTeacherByName(ctx, pe.Teacher)
			if err != nil {
				teacherID, err = s.teacherInitialsRepo.Upsert(ctx, teacherInitials, nil, group.InstitutionID)
				if err != nil {
					s.log.Errorf("ошибка поиска преподавателя: %v", err)
				}
			} else if t != nil {
				teacherID, err = s.teacherInitialsRepo.Upsert(ctx, teacherInitials, &t.ID, group.InstitutionID)
			}
		}

		entry := &domainSchedule.Schedule{
			GroupID:           groupID,
			SubjectID:         subj.ID,
			Date:              pe.Date,
			PairNumber:        pairNumber,
			Classroom:         pe.Classroom,
			TeacherInitialsID: &teacherID,
			StartTime:         combineTime(pe.Date, startTime),
			EndTime:           combineTime(pe.Date, endTime),
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
func (s *scheduleService) ByGroupID(ctx context.Context, groupID int) ([]*domainSchedule.Item, error) {
	entries, err := s.repo.ByGroupID(ctx, groupID)
	if err != nil {
		s.log.Errorf("Ошибка получения расписания: %v", err)
		return nil, err
	}
	var out []*domainSchedule.Item
	for _, e := range entries {
		var (
			tid      *int
			initials string
		)
		if e.TeacherInitialsID != nil {
			ti, err := s.teacherInitialsRepo.GetByID(ctx, *e.TeacherInitialsID)
			if err != nil {
				return nil, fmt.Errorf("initialsRepo.ByID: %w", err)
			}
			if ti != nil {
				tid = ti.TeacherID
				initials = ti.Initials
			}
		}

		out = append(out, &domainSchedule.Item{
			ID:              e.ID,
			GroupID:         e.GroupID,
			SubjectID:       e.SubjectID,
			Date:            e.Date,
			PairNumber:      e.PairNumber,
			Classroom:       e.Classroom,
			TeacherID:       tid,
			TeacherInitials: initials,
			StartTime:       e.StartTime,
			EndTime:         e.EndTime,
		})
	}
	return out, nil
}

func (s *scheduleService) updateInitials(ctx context.Context) error {
	initials, id, err := s.parserTeacher.FetchTeacher(context.Background())
	if err != nil {
		s.log.Errorf("Ошибка парсинга инициалов преподавателей: %v", err)
		return err
	}
	for _, initial := range initials {
		if strings.HasPrefix(initial, "_") {
			continue
		}
		_, err := s.teacherInitialsRepo.Upsert(ctx, initial, nil, id)
		if err != nil {
			s.log.Errorf("Ошибка сохраннеия инициала %v: %v", initial, err)
			continue
		}
	}
	return nil
}

func (s *scheduleService) Update(ctx context.Context, id int, req *dtoSchedule.UpdateScheduleReq) error {
	upd := make(map[string]interface{})
	if req.GroupId != nil {
		upd["group_id"] = *req.GroupId
	}
	if req.SubjectID != nil {
		upd["subject_id"] = *req.SubjectID
	}
	if req.Date != nil {
		upd["date"] = *req.Date
	}
	if req.PairNumber != nil {
		upd["pair_number"] = *req.PairNumber
	}
	if req.Classroom != nil {
		upd["classroom"] = *req.Classroom
	}
	if req.TeacherInitialsID != nil {
		upd["teacher_initials_id"] = *req.TeacherInitialsID
	}
	if req.StartTime != nil {
		upd["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		upd["end_time"] = *req.EndTime
	}
	if len(upd) == 0 {
		return nil
	}
	if err := s.repo.Update(ctx, id, upd); err != nil {
		s.log.Errorf("Ошибка: %v", err)
		return err
	}
	return nil
}

func (s *scheduleService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *scheduleService) ByID(ctx context.Context, id int) (*domainSchedule.Schedule, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *scheduleService) ByTeacherInitialsID(ctx context.Context, initialsID int) ([]*domainSchedule.Schedule, error) {
	entries, err := s.repo.ByTeacherInitialsID(ctx, initialsID)
	if err != nil {
		s.log.Errorf("Ошибка получения расписания по initials_id=%d: %v", initialsID, err)
		return nil, fmt.Errorf("не удалось получить расписание")
	}
	return entries, nil
}

// StartWorker запускает периодическое обновление расписания и инициалов.
func (s *scheduleService) StartWorker(interval time.Duration) {
	go func() {
		ctx := context.Background()
		for {

			groups, err := s.groupRepo.ByInstitutionID(ctx, s.parserSchedule.InstitutionID)
			if err != nil {
				s.log.Errorf("Ошибка получения групп: %v", err)
			}
			for _, group := range groups {
				err = s.Save(ctx, group.Name)
				if err != nil {
					s.log.Errorf("Ошибка обновления расписания: %v", err)
				} else {
					s.log.Info("Расписание успешно обновлено")
				}
				time.Sleep(time.Second * 10)
			}
			time.Sleep(interval)
		}
	}()
}

// StartWorkerInitials запускает периодическое обновление расписания и инициалов.
func (s *scheduleService) StartWorkerInitials(interval time.Duration) {
	go func() {
		for {
			err := s.updateInitials(context.Background())
			if err != nil {
				s.log.Errorf("Ошибка парсинга инициалов: %v", err)
			} else {
				s.log.Info("Инициалы успешно обновлены")
			}
			time.Sleep(interval)
		}
	}()
}
