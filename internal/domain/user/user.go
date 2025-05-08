package user

import (
	"errors"
	"strings"
)

// User представляет пользователя системы.
type User struct {
	ID           int
	Email        string
	PasswordHash []byte
	FullName     string
	IsTeacher    bool
	IsActive     bool
}

// CreateUser представляет пользователя системы.
type CreateUser struct {
	Email         string
	Password      string
	FullName      string
	IsTeacher     bool
	InstitutionID int
	GroupID       int
}

func (c *CreateUser) EmailMask() (string, error) {
	// Разделяем email по символу @
	parts := strings.Split(c.Email, "@")
	if len(parts) != 2 {
		return "", errors.New("некорректный формат email")
	}

	// Возвращаем доменную часть
	return parts[1], nil
}

// UpdateUser содержит поля, которые хотим обновить.
type UpdateUser struct {
	ID            int
	FullName      *string
	InstitutionID *int
	GroupID       *int // nil для преподавателей
	IsTeacher     bool
}

type Student struct {
	UserID        int
	InstitutionID int
	GroupID       int
}

type Teacher struct {
	UserID        int
	InstitutionID int
}

func (c *CreateUser) ConvertToUser(passwordHash []byte) *User {
	return &User{
		Email:        c.Email,
		PasswordHash: passwordHash,
		FullName:     c.FullName,
		IsTeacher:    c.IsTeacher,
	}
}
