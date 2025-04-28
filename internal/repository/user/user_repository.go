package user

import (
	domainUser "EduSync/internal/domain/user"
	"EduSync/internal/repository"
	"context"
	"database/sql"
	"fmt"
)

// userRepository обеспечивает работу с таблицей пользователей.
type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, tx *sql.Tx, user *domainUser.User) (int, error) {
	var userID int
	err := tx.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, full_name, is_teacher)
		VALUES ($1, $2, $3, $4) RETURNING id
	`, user.Email, user.PasswordHash, user.FullName, user.IsTeacher).Scan(&userID)
	return userID, err
}

func (r *userRepository) ByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	user := &domainUser.User{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, full_name, is_teacher 
		FROM users 
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.IsTeacher)
	if err == sql.ErrNoRows {
		return nil, nil // Пользователь не найден, возвращаем nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}
	return user, err
}

// BeginTx запускает транзакцию.
func (r *userRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *userRepository) ByID(ctx context.Context, ID int) (*domainUser.User, error) {
	user := &domainUser.User{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, full_name, is_teacher 
		FROM users 
		WHERE id = $1
	`, ID).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.IsTeacher)
	if err == sql.ErrNoRows {
		return nil, nil // Пользователь не найден, возвращаем nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}
	return user, err
}
