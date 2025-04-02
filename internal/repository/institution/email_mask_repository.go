package institution

import (
	domainInstitution "EduSync/internal/domain/institution"
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"
)

// emailMaskRepository реализует интерфейс EmailMaskRepository для PostgreSQL.
type emailMaskRepository struct {
	db *sql.DB
}

// NewEmailMaskRepository создает новый репозиторий почтовых масок.
func NewEmailMaskRepository(db *sql.DB) repository.EmailMaskRepository {
	return &emailMaskRepository{db: db}
}

// All возвращает все записи из таблицы institution_email_masks.
func (r *emailMaskRepository) All(ctx context.Context) ([]*domainInstitution.EmailMask, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT institution_id, email_mask
		FROM institution_email_masks
	`)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения почтовых масок: %w", err)
	}
	defer rows.Close()

	var masks []*domainInstitution.EmailMask
	for rows.Next() {
		mask := new(domainInstitution.EmailMask)
		if err := rows.Scan(&mask.InstitutionID, &mask.EmailMask); err != nil {
			return nil, fmt.Errorf("ошибка сканирования почтовой маски: %w", err)
		}
		masks = append(masks, mask)
	}
	return masks, nil
}

// ByEmailMask возвращает запись почтовой маски по значению email_mask.
func (r *emailMaskRepository) ByEmailMask(ctx context.Context, maskValue string) (*domainInstitution.EmailMask, error) {
	mask := new(domainInstitution.EmailMask)
	err := r.db.QueryRowContext(ctx, `
		SELECT institution_id, email_mask
		FROM institution_email_masks
		WHERE email_mask = $1
	`, maskValue).Scan(&mask.InstitutionID, &mask.EmailMask)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения почтовой маски: %w", err)
	}
	return mask, nil
}
