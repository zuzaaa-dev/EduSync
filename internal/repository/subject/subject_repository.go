package subject

import (
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainSubject "EduSync/internal/domain/subject"
)

// SubjectRepository управляет операциями с предметами в базе данных.
type subjectRepository struct {
	db *sql.DB
}

// NewSubjectRepository создает новый экземпляр репозитория предметов.
func NewSubjectRepository(db *sql.DB) repository.SubjectRepository {
	return &subjectRepository{db: db}
}

// Create создаёт предмет, если его ещё нет в БД.
func (r *subjectRepository) Create(ctx context.Context, name string, institutionID int) (int, error) {
	var existingID int
	err := r.db.QueryRowContext(ctx, `SELECT id FROM subjects WHERE name = $1 AND institution_id = $2`, name, institutionID).Scan(&existingID)
	if err == nil {
		// Предмет уже существует, возвращаем его ID
		return existingID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("ошибка проверки предмета: %w", err)
	}

	// Вставка нового предмета
	var subjectID int
	err = r.db.QueryRowContext(ctx, `INSERT INTO subjects (name, institution_id) VALUES ($1, $2) RETURNING id`, name, institutionID).Scan(&subjectID)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания предмета: %w", err)
	}
	return subjectID, nil
}

// GetByID получает предмет по его ID.
func (r *subjectRepository) GetByID(ctx context.Context, id int) (*domainSubject.Subject, error) {
	subject := &domainSubject.Subject{}
	err := r.db.QueryRowContext(ctx, `SELECT id, name, institution_id FROM subjects WHERE id = $1`, id).
		Scan(&subject.ID, &subject.Name, &subject.InstitutionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения предмета: %w", err)
	}
	return subject, nil
}

// GetByInstitutionID получает список предметов по ID учебного заведения.
func (r *subjectRepository) GetByInstitutionID(ctx context.Context, institutionID int) ([]*domainSubject.Subject, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, institution_id FROM subjects WHERE institution_id = $1`, institutionID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения предметов: %w", err)
	}
	defer rows.Close()

	var subjects []*domainSubject.Subject
	for rows.Next() {
		subj := &domainSubject.Subject{}
		if err := rows.Scan(&subj.ID, &subj.Name, &subj.InstitutionID); err != nil {
			return nil, fmt.Errorf("ошибка сканирования предмета: %w", err)
		}
		subjects = append(subjects, subj)
	}
	return subjects, nil
}

// GetByGroupID получает список предметов по ID группы.
func (r *subjectRepository) GetByGroupID(ctx context.Context, groupID int) ([]*domainSubject.Subject, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.id, s.name, s.institution_id 
		FROM subjects s
		JOIN schedule sc ON sc.subject_id = s.id
		WHERE sc.group_id = $1
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения предметов для группы: %w", err)
	}
	defer rows.Close()

	var subjects []*domainSubject.Subject
	for rows.Next() {
		subj := &domainSubject.Subject{}
		if err := rows.Scan(&subj.ID, &subj.Name, &subj.InstitutionID); err != nil {
			return nil, fmt.Errorf("ошибка сканирования предмета: %w", err)
		}
		subjects = append(subjects, subj)
	}
	return subjects, nil
}
