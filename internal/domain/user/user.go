package user

// User представляет пользователя системы.
type User struct {
	ID           int
	Email        string
	PasswordHash []byte
	FullName     string
	IsTeacher    bool
}

// CreateUser представляет пользователя системы.
type CreateUser struct {
	Email     string
	Password  string
	FullName  string
	IsTeacher bool
}

func (r *CreateUser) ConvertToUser(passwordHash *[]byte) *User {
	return &User{
		Email:        r.Email,
		PasswordHash: *passwordHash,
		FullName:     r.FullName,
		IsTeacher:    r.IsTeacher,
	}
}
