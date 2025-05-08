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
		SELECT id, email, password_hash, full_name, is_teacher, is_active
		FROM users 
		WHERE email = $1
	`, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.IsTeacher,
		&user.IsActive,
	)
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
		SELECT id, email, password_hash, full_name, is_teacher, is_active 
		FROM users 
		WHERE id = $1
	`, ID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.IsTeacher,
		&user.IsActive,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Пользователь не найден, возвращаем nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}
	return user, err
}
func (r *userRepository) Update(ctx context.Context, tx *sql.Tx, user *domainUser.User) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE users SET full_name = $1 WHERE id = $2`,
		user.FullName, user.ID,
	)
	return err
}

func (r *userRepository) Activate(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE users SET is_active = TRUE WHERE id = $1
    `, userID)
	return err
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID int, hashedPassword string) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE users SET password_hash = $1 WHERE id = $2
    `, hashedPassword, userID)
	return err
}
