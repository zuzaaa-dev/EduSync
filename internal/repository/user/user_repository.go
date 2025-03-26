package user

import (
	domainUser "EduSync/internal/domain/user"
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Repository обеспечивает работу с таблицей пользователей.
type Repository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user *domainUser.User) (int, error) {
	var userID int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, full_name, is_teacher)
		VALUES ($1, $2, $3, $4) RETURNING id
	`, user.Email, user.PasswordHash, user.FullName, user.IsTeacher).Scan(&userID)
	return userID, err
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	user := &domainUser.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, full_name, is_teacher 
		FROM users 
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.IsTeacher)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("неудалось получить пользователя: %w", err)
	}
	return user, err
}
