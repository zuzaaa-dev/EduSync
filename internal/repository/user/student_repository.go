package user

import (
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"

	domainUser "EduSync/internal/domain/user"
)

// studentRepository – конкретная реализация StudentRepository.
type studentRepository struct {
	db *sql.DB
}

// NewStudentRepository создаёт новый репозиторий для студентов.
func NewStudentRepository(db *sql.DB) repository.StudentRepository {
	return &studentRepository{db: db}
}

// CreateStudent добавляет запись в таблицу students.
func (r *studentRepository) CreateStudent(ctx context.Context, userID, institutionID, groupID int) error {
	var err error
	if groupID > 0 {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO students (user_id, institution_id, group_id)
			VALUES ($1, $2, $3)
		`, userID, institutionID, groupID)
	} else {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO students (user_id, institution_id)
			VALUES ($1, $2)
		`, userID, institutionID)
	}
	if err != nil {
		return fmt.Errorf("ошибка при создании студента: %v", err)
	}
	return nil
}

// GetStudentByUserID получает студента по user_id.
func (r *studentRepository) GetStudentByUserID(ctx context.Context, userID int) (*domainUser.Student, error) {
	student := &domainUser.Student{}
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, institution_id, group_id FROM students WHERE user_id = $1
	`, userID).Scan(&student.UserID, &student.InstitutionID, &student.GroupID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении студента по user_id: %v", err)
	}
	return student, nil
}

// GetStudentsByGroupID получает список студентов по group_id.
func (r *studentRepository) GetStudentsByGroupID(ctx context.Context, groupID int) ([]*domainUser.Student, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, institution_id, group_id FROM students WHERE group_id = $1
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении студентов по group_id: %v", err)
	}
	defer rows.Close()

	var students []*domainUser.Student
	for rows.Next() {
		student := &domainUser.Student{}
		if err := rows.Scan(&student.UserID, &student.InstitutionID, &student.GroupID); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании студента: %v", err)
		}
		students = append(students, student)
	}
	return students, nil
}
