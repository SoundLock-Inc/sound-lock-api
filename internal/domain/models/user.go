package models

import "github.com/google/uuid"

type User struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

type UserDTO struct {
	UUID     uuid.UUID
	Email    string
	Password string
}

func (u *UserDTO) ToUser(user *User) {
	user.ID = u.UUID
	user.Email = u.Email
}
