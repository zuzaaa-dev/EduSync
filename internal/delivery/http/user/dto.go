package user

import (
	domainUser "EduSync/internal/domain/user"
	"errors"
)

type RegistrationUserReq struct {
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required"`
	FullName      string `json:"full_name" binding:"required"`
	InstitutionID int    `json:"institution_id" binding:"required"`
	GroupID       int    `json:"group_id"`
	IsTeacher     *bool  `json:"is_teacher" binding:"required"`
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

type LoginUserReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type PairTokenResp struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}
