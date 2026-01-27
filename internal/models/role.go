package models

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleUser
}
