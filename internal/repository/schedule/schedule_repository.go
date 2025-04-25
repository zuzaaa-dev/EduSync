package schedule

import (
	domainSchedule "EduSync/internal/domain/schedule"
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// PostgresScheduleRepository реализует интерфейс Repository для расписания.
type PostgresScheduleRepository struct {
	db *sql.DB
}

// NewScheduleRepository создает новый репозиторий расписания.
func NewScheduleRepository(db *sql.DB) repository.ScheduleRepository {
	return &PostgresScheduleRepository{db: db}
}

// Save сохраняет записи расписания в БД.
// Здесь можно использовать транзакцию для пакетной вставки.
func (r *PostgresScheduleRepository) Save(ctx context.Context, entries []*domainSchedule.Schedule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO schedule (group_id, subject_id, date, pair_number, classroom, teacher_initials_id, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("не удалось подготовить запрос: %v", err)
	}
	defer stmt.Close()

	var teacherInitID *int
	for _, entry := range entries {
		if *entry.TeacherInitialsID != 0 {
			teacherInitID = entry.TeacherInitialsID
		} else {
			teacherInitID = nil
		}
		_, err := stmt.ExecContext(ctx,
			entry.GroupID,
			entry.SubjectID,
			entry.Date,
			entry.PairNumber,
			entry.Classroom,
			teacherInitID,
			entry.StartTime,
			entry.EndTime,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка при сохранении расписания: %v", err)
		}
	}

	return tx.Commit()
}

func (r *PostgresScheduleRepository) Create(ctx context.Context, s *domainSchedule.Schedule) (int, error) {
	var id int
	query := `
		INSERT INTO schedule (group_id, subject_id, date, pair_number, classroom, teacher_initials_id, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		s.GroupID, s.SubjectID, s.Date, s.PairNumber, s.Classroom,
		s.TeacherInitialsID, s.StartTime, s.EndTime,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания записи расписания: %w", err)
	}
	return id, nil
}

// ByGroupID возвращает расписание для заданной группы.
func (r *PostgresScheduleRepository) ByGroupID(ctx context.Context, groupID int) ([]*domainSchedule.Schedule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, group_id, subject_id, date, pair_number, classroom, teacher_initials_id, start_time, end_time
		FROM schedule
		WHERE group_id = $1
		ORDER BY date, pair_number
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения расписания: %v", err)
	}
	defer rows.Close()

	var entries []*domainSchedule.Schedule
	for rows.Next() {
		entry := &domainSchedule.Schedule{}
		var teacherInitialsID *int

		err := rows.Scan(&entry.ID, &entry.GroupID, &entry.SubjectID, &entry.Date, &entry.PairNumber, &entry.Classroom,
			&teacherInitialsID, &entry.StartTime, &entry.EndTime)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи расписания: %v", err)
		}
		if teacherInitialsID != nil {
			entry.TeacherInitialsID = teacherInitialsID
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *PostgresScheduleRepository) Update(ctx context.Context, id int, upd map[string]interface{}) error {
	if len(upd) == 0 {
		return nil
	}
	sets, args := []string{}, []interface{}{}
	i := 1
	for col, val := range upd {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	// последний аргумент — id
	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE schedule SET %s WHERE id = $%d",
		strings.Join(sets, ", "),
		i,
	)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresScheduleRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM schedule WHERE id = $1", id)
	return err
}

func (r *PostgresScheduleRepository) GetByID(ctx context.Context, id int) (*domainSchedule.Schedule, error) {
	row := r.db.QueryRowContext(ctx, `
      SELECT id, group_id, subject_id, date, pair_number, classroom,
             teacher_initials_id, start_time, end_time
        FROM schedule WHERE id = $1`, id)
	s := new(domainSchedule.Schedule)
	var tiid sql.NullInt64
	if err := row.Scan(&s.ID, &s.GroupID, &s.SubjectID, &s.Date, &s.PairNumber,
		&s.Classroom, &tiid, &s.StartTime, &s.EndTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if tiid.Valid {
		v := int(tiid.Int64)
		s.TeacherInitialsID = &v
	}
	return s, nil
}

func (r *PostgresScheduleRepository) ByTeacherInitialsID(ctx context.Context, initialsID int) ([]*domainSchedule.Schedule, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, group_id, subject_id, date, pair_number, classroom, teacher_initials_id, start_time, end_time
        FROM schedule
        WHERE teacher_initials_id = $1
        ORDER BY date, pair_number
    `, initialsID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения расписания по initials_id=%d: %w", initialsID, err)
	}
	defer rows.Close()

	var entries []*domainSchedule.Schedule
	for rows.Next() {
		ent := &domainSchedule.Schedule{}
		var tiID sql.NullInt64
		if err := rows.Scan(
			&ent.ID,
			&ent.GroupID,
			&ent.SubjectID,
			&ent.Date,
			&ent.PairNumber,
			&ent.Classroom,
			&tiID,
			&ent.StartTime,
			&ent.EndTime,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи расписания: %w", err)
		}
		entries = append(entries, ent)
	}
	return entries, nil
}
