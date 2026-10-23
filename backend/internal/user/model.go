package user

import "time"

type Role string

const (
	RoleOrganizer Role = "organizer"
	RoleVisitor   Role = "visitor"
)

func (r Role) IsValid() bool {
	return r == RoleOrganizer || r == RoleVisitor
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CreateUserInput struct {
	Email        string
	PasswordHash string
	Role         Role
}
