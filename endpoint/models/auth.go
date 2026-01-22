package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Users struct {
	ID        int        `json:"id"`
	Role      string     `json:"role"`
	Nip       string     `json:"nip"`
	Email     string     `json:"email"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	CreatedBy *string    `json:"created_by"`
	UpdatedAt *time.Time `json:"updated_at"`
	UpdatedBy *string    `json:"updated_by"`
}

type LoginRequest struct {
	Nip      string `json:"nip"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User  Users  `json:"user"`
	Token string `json:"token"`
}

type JWTClaims struct {
	UserID int    `json:"user_id"`
	Nip    string `json:"nip"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
