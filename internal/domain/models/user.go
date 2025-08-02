package models

import "github.com/google/uuid"

type User struct {
	Email    string `json:"email"`
	Password string
}

type UserDTO struct {
	UUID     uuid.UUID
	Email    string
	Password string
}
