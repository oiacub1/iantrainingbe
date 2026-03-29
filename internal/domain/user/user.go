package user

import "time"

type UserRole string

const (
	RoleTrainer UserRole = "TRAINER"
	RoleStudent UserRole = "STUDENT"
	RoleAdmin   UserRole = "ADMIN"
)

type UserStatus string

const (
	StatusActive   UserStatus = "ACTIVE"
	StatusInactive UserStatus = "INACTIVE"
	StatusSuspended UserStatus = "SUSPENDED"
)

type User struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Phone     string     `json:"phone"`
	Role      UserRole   `json:"role"`
	Status    UserStatus `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (r UserRole) IsValid() bool {
	switch r {
	case RoleTrainer, RoleStudent, RoleAdmin:
		return true
	default:
		return false
	}
}

func (s UserStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusSuspended:
		return true
	default:
		return false
	}
}
