package user

import "github.com/golang-jwt/jwt/v5"

// TokenClaims — данные, которые будут помещаться в JWT.
type TokenClaims struct {
	ID            int    `json:"id"`
	IsTeacher     bool   `json:"is_teacher"`
	Email         string `json:"email"`
	FullName      string `json:"full_name"`
	InstitutionId int    `json:"institution_id"`
	GroupId       int    `json:"group_id"`
	jwt.RegisteredClaims
}
