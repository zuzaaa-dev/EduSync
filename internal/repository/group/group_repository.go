package group

import (
	domainGroup "EduSync/internal/domain/group"
	"database/sql"
	"fmt"
)

// GroupRepository реализует интерфейс Repository для PostgreSQL.
type GroupRepository struct {
	db *sql.DB
}

// NewGroupRepository возвращает новую реализацию Repository.
func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// SaveGroups сохраняет группы в БД. Здесь можно реализовать логику обновления (UPSERT) или простое добавление.
func (r *GroupRepository) SaveGroups(groups []*domainGroup.Group) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO groups (name, institution_id)
		VALUES ($1, $2)
		ON CONFLICT (name, institution_id) DO NOTHING
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("не удалось подготовить запрос: %v", err)
	}
	defer stmt.Close()

	for _, g := range groups {
		_, err := stmt.Exec(g.Name, g.InstitutionID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка при сохранении группы %s: %v", g.Name, err)
		}
	}

	return tx.Commit()
}

func (r *GroupRepository) GetByInstitutionID(institutionID int) ([]*domainGroup.Group, error) {
	rows, err := r.db.Query(`SELECT id, name, institution_id FROM groups WHERE institution_id = $1`, institutionID)
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
