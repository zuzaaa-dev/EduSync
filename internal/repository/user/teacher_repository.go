package user

import (
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"strings"

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

// GetTeachersBySurname возвращает преподавателей, у которых фамилия начинается с заданного значения.
func (r *teacherRepository) GetTeachersBySurname(ctx context.Context, surname string) ([]*domainUser.User, error) {
	// Предполагается, что полное имя хранится как "Фамилия Имя Отчество"
	// Используем ILIKE для нечувствительности к регистру
	query := `
		SELECT id, email, password_hash, full_name, is_teacher
		FROM users
		WHERE is_teacher = TRUE AND full_name ILIKE $1
	`
	// Формируем шаблон: "Фамилия %"
	pattern := surname + " %"
	rows, err := r.db.QueryContext(ctx, query, pattern)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса преподавателей: %w", err)
	}
	defer rows.Close()

	var teachers []*domainUser.User
	for rows.Next() {
		teacher := new(domainUser.User)
		err := rows.Scan(&teacher.ID, &teacher.Email, &teacher.PasswordHash, &teacher.FullName, &teacher.IsTeacher)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования преподавателя: %w", err)
		}
		// Дополнительно убедимся, что фамилия точно совпадает (если, например, у кого-то фамилия похожая)
		fullNameParts := strings.Fields(teacher.FullName)
		if len(fullNameParts) > 0 && strings.EqualFold(fullNameParts[0], surname) {
			teachers = append(teachers, teacher)
		}
	}

	return teachers, nil
}
