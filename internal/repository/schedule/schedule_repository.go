package schedule

import (
	domainSchedule "EduSync/internal/domain/schedule"
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"
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
		INSERT INTO schedule (group_id, subject_id, date, pair_number, classroom, teacher_id, teacher_initials, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT DO NOTHING
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("не удалось подготовить запрос: %v", err)
	}
	defer stmt.Close()

	for _, entry := range entries {
		_, err := stmt.ExecContext(ctx,
			entry.GroupID,
			entry.SubjectID,
			entry.Date,
			entry.PairNumber,
			entry.Classroom,
			entry.TeacherID,
			entry.TeacherInitials,
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

// ByGroupID возвращает расписание для заданной группы.
func (r *PostgresScheduleRepository) ByGroupID(ctx context.Context, groupID int) ([]*domainSchedule.Schedule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, group_id, subject_id, date, pair_number, classroom, teacher_id, teacher_initials, start_time, end_time
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
		var teacherID sql.NullInt64
		var teacherInitials sql.NullString

		err := rows.Scan(&entry.ID, &entry.GroupID, &entry.SubjectID, &entry.Date, &entry.PairNumber, &entry.Classroom,
			&teacherID, &teacherInitials, &entry.StartTime, &entry.EndTime)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи расписания: %v", err)
		}
		if teacherID.Valid {
			id := int(teacherID.Int64)
			entry.TeacherID = &id
		} else {
			entry.TeacherID = nil
		}
		if teacherInitials.Valid {
			entry.TeacherInitials = teacherInitials.String
		} else {
			entry.TeacherInitials = ""
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
