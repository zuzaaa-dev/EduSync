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
	"sort"
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

func (s *scheduleService) Save(ctx context.Context, groupName string) error {
	s.log.Infof("Сохраняем расписание для группы %q", groupName)

	// 1) Парсим сырое расписание
	parsed, err := s.parserSchedule.FetchSchedule(ctx, groupName)
	if err != nil {
		s.log.Errorf("парсер вернул ошибку: %v", err)
		return fmt.Errorf("не удалось распарсить расписание")
	}

	// 2) Получаем ID группы и института
	group, err := s.groupRepo.ByName(ctx, groupName)
	if err != nil {
		s.log.Errorf("ошибка поиска группы %q: %v", groupName, err)
		return fmt.Errorf("группа не найдена")
	}

	instID := group.InstitutionID
	groupID := group.ID

	// 3) Для каждой записи раскладываем Start/End в time.Time и собираем по дате
	type raw struct {
		pe    schedule.ScheduleEntry // ваша внутренняя модель parser.FetchSchedule
		start time.Time
		end   time.Time
	}

	byDate := make(map[time.Time][]raw)
	for _, pe := range parsed {
		st, err := time.Parse("15:04", pe.StartTime)
		if err != nil {
			s.log.Warnf("некоректное время пары %q: %v", pe.StartTime, err)
			continue
		}
		en, err := time.Parse("15:04", pe.EndTime)
		if err != nil {
			s.log.Warnf("некоректное время конца пары %q: %v", pe.EndTime, err)
			continue
		}
		// обнуляем час/минута/секунды даты — чтобы ключом была ровно дата
		d := pe.Date.Truncate(24 * time.Hour)
		byDate[d] = append(byDate[d], raw{pe: pe, start: st, end: en})
	}

	// 4) Собираем финальный слайс domainSchedule.Schedule с правильно пронумерованными парами
	var toSave []*domainSchedule.Schedule
	for d, raws := range byDate {
		// сортируем по start asc
		sort.Slice(raws, func(i, j int) bool {
			return raws[i].start.Before(raws[j].start)
		})

		for idx, r := range raws {
			// idx==0 => pairNumber=1, и т.д.
			pairNum := idx + 1

			// subject: ищем или создаём
			subj, err := s.subjectSvc.ByNameAndInstitution(ctx, r.pe.Discipline, instID)
			if err != nil {
				s.log.Errorf("subjectSvc.ByName %q: %v", r.pe.Discipline, err)
				continue
			}
			if subj == nil {
				id, err := s.subjectSvc.Create(ctx, r.pe.Discipline, instID)
				if err != nil {
					s.log.Errorf("subjectSvc.Create %q: %v", r.pe.Discipline, err)
					continue
				}
				subj, _ = s.subjectSvc.ByID(ctx, id)
			}

			// teacher_initials upsert
			var tiID *int
			initials := r.pe.Teacher
			if initials != "" && initials != "-" {
				// если нашли реального учителя
				usr, err := s.userSvc.FindTeacherByName(ctx, initials)
				var teacherPtr *int
				if err == nil && usr != nil {
					teacherPtr = &usr.ID
				}
				id, err := s.teacherInitialsRepo.Upsert(ctx, initials, teacherPtr, instID)
				if err != nil {
					s.log.Errorf("teacherInitialsRepo.Upsert %q: %v", initials, err)
				} else {
					tiID = &id
				}
			}

			toSave = append(toSave, &domainSchedule.Schedule{
				GroupID:           groupID,
				SubjectID:         subj.ID,
				Date:              d,
				PairNumber:        pairNum,
				Classroom:         r.pe.Classroom,
				TeacherInitialsID: tiID,
				StartTime:         combineTime(d, r.start),
				EndTime:           combineTime(d, r.end),
			})
		}
	}

	// 5) Сохраняем всё пачкой
	if err := s.repo.Save(ctx, toSave); err != nil {
		s.log.Errorf("repo.Save: %v", err)
		return fmt.Errorf("не удалось сохранить расписание")
	}

	s.log.Infof("успешно сохранено %d записей", len(toSave))
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

	subs, err := s.subjectSvc.ByGroupID(ctx, groupID)
	if err != nil {
		s.log.Errorf("Ошибка получения предметов для группы %d: %v", groupID, err)
		return nil, fmt.Errorf("не удалось получить предметы")
	}
	subjectMap := make(map[int]string, len(subs))
	for _, sub := range subs {
		subjectMap[sub.ID] = sub.Name
	}

	var out []*domainSchedule.Item
	for _, e := range entries {
		// вытаскиваем информацию по преподавателю/инициалам
		var (
			tid      *int
			initials string
		)
		if e.TeacherInitialsID != nil {
			ti, err := s.teacherInitialsRepo.GetByID(ctx, *e.TeacherInitialsID)
			if err != nil {
				return nil, fmt.Errorf("initialsRepo.GetByID: %w", err)
			}
			if ti != nil {
				tid = ti.TeacherID
				initials = ti.Initials
			}
		}

		subjName := subjectMap[e.SubjectID]

		out = append(out, &domainSchedule.Item{
			ID:              e.ID,
			GroupID:         e.GroupID,
			Subject:         subjName,
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

func (s *scheduleService) ByTeacherInitialsID(ctx context.Context, initialsID int) ([]*domainSchedule.Item, error) {
	entries, err := s.repo.ByTeacherInitialsID(ctx, initialsID)
	if err != nil {
		s.log.Errorf("Ошибка получения расписания по initials_id=%d: %v", initialsID, err)
		return nil, fmt.Errorf("не удалось получить расписание")
	}

	subjectIDs := make(map[int]struct{}, len(entries))
	for _, e := range entries {
		subjectIDs[e.SubjectID] = struct{}{}
	}

	subjMap := make(map[int]string, len(subjectIDs))
	for id := range subjectIDs {
		sub, err := s.subjectSvc.ByID(ctx, id)
		if err != nil {
			s.log.Errorf("Ошибка получения предмета %d: %v", id, err)
			continue
		}
		if sub != nil {
			subjMap[id] = sub.Name
		}
	}

	var out []*domainSchedule.Item
	for _, e := range entries {
		subjectName := subjMap[e.SubjectID]

		var (
			teacherID   *int
			initialsStr string
		)
		if e.TeacherInitialsID != nil {
			ti, err := s.teacherInitialsRepo.GetByID(ctx, *e.TeacherInitialsID)
			if err != nil {
				s.log.Errorf("initialsRepo.GetByID(%d): %v", *e.TeacherInitialsID, err)
			} else if ti != nil {
				initialsStr = ti.Initials
				teacherID = ti.TeacherID
			}
		}

		out = append(out, &domainSchedule.Item{
			ID:              e.ID,
			GroupID:         e.GroupID,
			Subject:         subjectName,
			Date:            e.Date,
			PairNumber:      e.PairNumber,
			Classroom:       e.Classroom,
			TeacherID:       teacherID,
			TeacherInitials: initialsStr,
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
	if req.GroupID != nil {
		upd["group_id"] = *req.GroupID
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

func (s *scheduleService) Create(ctx context.Context, req *dtoSchedule.CreateScheduleReq) (int, error) {
	entry := &domainSchedule.Schedule{
		GroupID:           req.GroupID,
		SubjectID:         req.SubjectID,
		Date:              req.Date,
		PairNumber:        req.PairNumber,
		Classroom:         req.Classroom,
		TeacherInitialsID: req.TeacherInitialsID,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
	}
	return s.repo.Create(ctx, entry)
}

func (s *scheduleService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *scheduleService) ByID(ctx context.Context, id int) (*domainSchedule.Schedule, error) {
	return s.repo.GetByID(ctx, id)
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
