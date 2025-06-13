package models

import (
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID          string `json:"user_id"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	PengurusLevel   string `json:"pengurus_level"`
	PengurusJabatan string `json:"pengurus_jabatan"`
	jwt.RegisteredClaims
}
