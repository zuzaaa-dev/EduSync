// internal/repository/schedule/teacher_initials_repository.go
package schedule

import (
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"

	domain "EduSync/internal/domain/schedule"
)

type teacherInitialsRepo struct {
	db *sql.DB
}

func NewTeacherInitialsRepository(db *sql.DB) repository.TeacherInitialsRepository {
	return &teacherInitialsRepo{db: db}
}

func (r *teacherInitialsRepo) Upsert(ctx context.Context, initials string, teacherID *int, institutionID int) (int, error) {
	const q = `
    INSERT INTO teacher_initials (initials, teacher_id, institution_id)
    VALUES ($1, $2, $3)
    ON CONFLICT (initials, institution_id) DO UPDATE
      SET teacher_id = COALESCE(teacher_initials.teacher_id, EXCLUDED.teacher_id)
    RETURNING id
    `

	// Если teacherID == nil или == 0, передаём в БД NULL
	var tid interface{}
	if teacherID != nil && *teacherID > 0 {
		tid = *teacherID
	} else {
		tid = nil
	}

	var id int
	err := r.db.QueryRowContext(ctx, q, initials, tid, institutionID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert teacher_initials: %w", err)
	}
	return id, nil
}

func (r *teacherInitialsRepo) GetByInitials(ctx context.Context, initials string, institutionID int) (*domain.TeacherInitials, error) {
	const q = `
    SELECT id, initials, teacher_id, institution_id
      FROM teacher_initials
     WHERE initials = $1 AND institution_id = $2
    `
	row := r.db.QueryRowContext(ctx, q, initials, institutionID)
	rec := &domain.TeacherInitials{}
	err := row.Scan(&rec.ID, &rec.Initials, &rec.TeacherID, &rec.InstitutionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select teacher_initials: %w", err)
	}
	return rec, nil
}

func (r *teacherInitialsRepo) GetByID(ctx context.Context, id int) (*domain.TeacherInitials, error) {
	const q = `
      SELECT id, initials, teacher_id, institution_id
        FROM teacher_initials
       WHERE id = $1`
	ti := &domain.TeacherInitials{}
	err := r.db.QueryRowContext(ctx, q, id).
		Scan(&ti.ID, &ti.Initials, &ti.TeacherID, &ti.InstitutionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ByID teacher_initials: %w", err)
	}
	return ti, nil
}

func (r *teacherInitialsRepo) GetAll(ctx context.Context, institutionID int) ([]*domain.TeacherInitials, error) {
	const q = `
      SELECT id, initials, teacher_id, institution_id
        FROM teacher_initials
        WHERE institution_id = $1`
	rows, err := r.db.QueryContext(ctx, q, institutionID)
	if err != nil {
		return nil, fmt.Errorf("teacher_initials.GetAll: %w", err)
	}
	defer rows.Close()

	var list []*domain.TeacherInitials
	for rows.Next() {
		ti := &domain.TeacherInitials{}
		if err := rows.Scan(&ti.ID, &ti.Initials, &ti.TeacherID, &ti.InstitutionID); err != nil {
			return nil, fmt.Errorf("teacher_initials scan: %w", err)
		}
		list = append(list, ti)
	}
	return list, nil
}
