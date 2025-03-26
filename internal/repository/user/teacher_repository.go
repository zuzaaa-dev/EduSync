package user

import (
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"

	domainUser "EduSync/internal/domain/user"
)

// teacherRepository – конкретная реализация TeacherRepository.
type teacherRepository struct {
	db *sql.DB
}

// NewTeacherRepository создаёт новый репозиторий для преподавателей.
func NewTeacherRepository(db *sql.DB) repository.TeacherRepository {
	return &teacherRepository{db: db}
}

// CreateTeacher добавляет запись в таблицу teachers.
func (r *teacherRepository) CreateTeacher(ctx context.Context, userID, institutionID int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO teachers (user_id, institution_id)
		VALUES ($1, $2)
	`, userID, institutionID)
	if err != nil {
		return fmt.Errorf("ошибка при создании преподавателя: %v", err)
	}
	return nil
}

// GetTeacherByUserID получает преподавателя по user_id.
func (r *teacherRepository) GetTeacherByUserID(ctx context.Context, userID int) (*domainUser.Teacher, error) {
	teacher := &domainUser.Teacher{}
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, institution_id FROM teachers WHERE user_id = $1
	`, userID).Scan(&teacher.UserID, &teacher.InstitutionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении преподавателя по user_id: %v", err)
	}
	return teacher, nil
}

// GetTeachersByInstitutionID получает список преподавателей по institution_id.
func (r *teacherRepository) GetTeachersByInstitutionID(ctx context.Context, institutionID int) ([]*domainUser.Teacher, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, institution_id FROM teachers WHERE institution_id = $1
	`, institutionID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении преподавателей по institution_id: %v", err)
	}
	defer rows.Close()

	var teachers []*domainUser.Teacher
	for rows.Next() {
		teacher := &domainUser.Teacher{}
		if err := rows.Scan(&teacher.UserID, &teacher.InstitutionID); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании преподавателя: %v", err)
		}
		teachers = append(teachers, teacher)
	}
	return teachers, nil
}
