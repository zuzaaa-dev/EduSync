package institution

import (
	domainInstitution "EduSync/internal/domain/institution"
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Repository – репозиторий заведений.
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) repository.InstitutionRepository {
	return &Repository{db: db}
}

func (r *Repository) ByID(ctx context.Context, id int) (*domainInstitution.Institution, error) {
	inst := &domainInstitution.Institution{}
	err := r.db.QueryRowContext(ctx, "SELECT id, name FROM institutions WHERE id = $1", id).
		Scan(&inst.ID, &inst.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения учреждения по id: %v", err)
	}
	return inst, nil
}

func (r *Repository) All(ctx context.Context) ([]*domainInstitution.Institution, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM institutions")
	if err != nil {
		return nil, fmt.Errorf("ошибка получения учреждений: %v", err)
	}
	defer rows.Close()

	var institutions []*domainInstitution.Institution
	for rows.Next() {
		inst := &domainInstitution.Institution{}
		if err := rows.Scan(&inst.ID, &inst.Name); err != nil {
			return nil, fmt.Errorf("ошибка сканирования учреждения: %v", err)
		}
		institutions = append(institutions, inst)
	}
	return institutions, nil
}
