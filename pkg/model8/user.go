package model8

import (
	"github.com/gofrs/uuid/v5"
)

type Roletype string

const (
	RoleUser  Roletype = "user"  // standard
	RoleAdmin Roletype = "admin" // Admin
)

type User8 struct {
	Id     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Role   Roletype  `json:"role"`
	Report bool      `json:"report"`
}
