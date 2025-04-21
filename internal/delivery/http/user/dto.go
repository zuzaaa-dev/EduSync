package user

import (
	domainUser "EduSync/internal/domain/user"
	"errors"
)

// LoginUserReq тело запроса на вход
// swagger:model
type LoginUserReq struct {
	// Email пользователя
	// required: true
	// example: user@example.com
	Email string `json:"email" binding:"required,email" example:"student@gmail.com"`
	// Пароль
	// required: true
	// example: mysecretpassword
	Password string `json:"password" binding:"required" example:"mysecretpassword"`
}

// RegistrationUserReq — тело запроса на регистрацию.
// swagger:model
type RegistrationUserReq struct {
	// Email пользователя
	// required: true
	Email string `json:"email" binding:"required,email" example:"student@gmail.com"`
	// Пароль
	// required: true
	// example: mysecretpassword
	Password string `json:"password" binding:"required" example:"mysecretpassword"`
	// Полное имя пользователя
	// required: true
	FullName string `json:"full_name" binding:"required" example:"Фамилия Имя Отчество"`
	// ID учебного учреждения
	// required: true
	InstitutionID int `json:"institution_id" binding:"required" example:"1"`
	// ID группы
	GroupID int `json:"group_id" example:"1"`
	// Является ли пользователь учителем
	// required: true
	IsTeacher *bool `json:"is_teacher" binding:"required" example:"false"`
}

func (r *RegistrationUserReq) ConvertToSvc() (domainUser.CreateUser, error) {
	userData := domainUser.CreateUser{
		Email:         r.Email,
		Password:      r.Password,
		FullName:      r.FullName,
		IsTeacher:     *r.IsTeacher,
		InstitutionID: r.InstitutionID,
	}

	if !*r.IsTeacher {
		if r.GroupID <= 0 {
			return domainUser.CreateUser{}, errors.New("обязательное поле group_id не задано")
		}
		userData.GroupID = r.GroupID
	}
	return userData, nil
}

func NewRegistrationUserReq(email string, password string, fullName string, isTeacher bool) *RegistrationUserReq {
	return &RegistrationUserReq{Email: email, Password: password, FullName: fullName, IsTeacher: &isTeacher}
}

// RefreshTokenReq тело запроса обновления токена
// swagger:model
type RefreshTokenReq struct {
	// Refresh токен
	// required: true
	// example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ErrorResponse тело запроса обновления токена
// swagger:model
type ErrorResponse struct {
	// Refresh токен

	// required: true
	// example: Ошибка
	Error string `json:"error" example:"Ошибка"`
}

// PairTokenResp — ответ с парами токенов.
// swagger:model
type PairTokenResp struct {
	// JWT для доступа
	AccessToken string `json:"access_token" binding:"required"`
	// JWT для обновления
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ProfileResp struct {
	// ID пользователя
	// required: true
	UserID int `json:"user_id" example:"1"`
	// Email пользователя
	// required: true
	Email string `json:"email" binding:"required,email" example:"student@gmail.com"`
	// Полное имя пользователя
	// required: true
	FullName string `json:"full_name" binding:"required" example:"Фамилия Имя Отчество"`
	// ID учебного учреждения
	// required: true
	InstitutionID int `json:"institution_id" binding:"required" example:"1"`
	// ID группы
	GroupID int `json:"group_id" example:"1"`
	// Является ли пользователь учителем
	// required: true
	IsTeacher bool `json:"is_teacher" binding:"required" example:"false"`
}
