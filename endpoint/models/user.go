package models

import "time"

type User struct {
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

type CreateUserRequest struct {
	Nip   string `json:"nip"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateUserRequest struct {
	Nip   string `json:"nip"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
