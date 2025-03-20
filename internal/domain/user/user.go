package user

// User представляет пользователя системы.
type User struct {
	ID           int
	Email        string
	PasswordHash string
	FullName     string
	IsTeacher    bool
}

// Repository описывает контракт для работы с пользователями.
type Repository interface {
	CreateUser(email, passwordHash, fullName string, isTeacher bool) (int, error)
	GetUserByEmail(email string) (*User, error)
}
