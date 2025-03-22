package user

import domainUser "EduSync/internal/domain/user"

type RegistrationUserReq struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	FullName  string `json:"full_name" binding:"required"`
	IsTeacher bool   `json:"is_teacher" binding:"required"`
}

func (r *RegistrationUserReq) ConvertToSvc() domainUser.CreateUser {
	return domainUser.CreateUser{
		Email:     r.Email,
		Password:  r.Password,
		FullName:  r.FullName,
		IsTeacher: r.IsTeacher,
	}
}

func NewRegistrationUserReq(email string, password string, fullName string, isTeacher bool) *RegistrationUserReq {
	return &RegistrationUserReq{Email: email, Password: password, FullName: fullName, IsTeacher: isTeacher}
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
