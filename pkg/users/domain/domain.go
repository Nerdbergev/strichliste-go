package domain

import "time"

type User struct {
	ID         int64
	Name       string
	Email      *string
	Balance    int64
	IsDisabled bool
	IsActive   bool
	Created    time.Time
	Updated    *time.Time
}

type UserRepository interface {
	GetAll() ([]User, error)
	FindActive(time.Time) ([]User, error)
	FindInactive(time.Time) ([]User, error)
	FindById(int64) (User, error)
	StoreUser(User) (User, error)
	UpdateUser(User) error
	DeleteById(int64) error
}
