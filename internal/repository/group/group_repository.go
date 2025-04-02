package group

import (
	domainGroup "EduSync/internal/domain/group"
	"EduSync/internal/repository"
	"context"
	"database/sql"
)

// GroupRepository реализует интерфейс Repository для PostgreSQL.
type GroupRepository struct {
	db *sql.DB
}

// NewGroupRepository возвращает новую реализацию Repository.
func NewGroupRepository(db *sql.DB) repository.GroupRepository {
	return &GroupRepository{db: db}
}

// SaveGroups сохраняет группы в БД. Здесь можно реализовать логику обновления (UPSERT) или простое добавление.
func (r *GroupRepository) Save(ctx context.Context, groups []*domainGroup.Group) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO groups (name, institution_id)
		VALUES ($1, $2)
		ON CONFLICT (name, institution_id) DO NOTHING
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, g := range groups {
		_, err := stmt.Exec(g.Name, g.InstitutionID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *GroupRepository) ByInstitutionID(ctx context.Context, institutionID int) ([]*domainGroup.Group, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, institution_id 
		FROM groups 
		WHERE institution_id = $1`, institutionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domainGroup.Group
	for rows.Next() {
		g := new(domainGroup.Group)
		if err := rows.Scan(&g.ID, &g.Name, &g.InstitutionID); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *GroupRepository) ById(ctx context.Context, id int) (*domainGroup.Group, error) {
	group := &domainGroup.Group{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, institution_id 
		FROM groups 
		WHERE id = $1`, id).Scan(&group.ID, &group.Name, &group.InstitutionID)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (r *GroupRepository) ByName(ctx context.Context, name string) (*domainGroup.Group, error) {
	group := &domainGroup.Group{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, institution_id 
		FROM groups 
		WHERE name = $1`, name).Scan(&group.ID, &group.Name, &group.InstitutionID)
	if err != nil {
		return nil, err
	}

	return group, nil
}
