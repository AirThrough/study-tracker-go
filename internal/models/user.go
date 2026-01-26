package models

import "time"

type User struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
