package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrPersistanceError = errors.New("persistance error")
)

type UserNotFoundError struct {
	UID   int64
	Cause error
}

func (e UserNotFoundError) Error() string {
	return fmt.Sprintf("User '%d' not found", e.UID)
}

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

func (u *User) AddBalance(amount int64) {
	u.Balance += amount
	now := time.Now()
	u.Updated = &now
}

type UserRepository interface {
	GetAll() ([]User, error)
	FindActive(time.Time) ([]User, error)
	FindInactive(time.Time) ([]User, error)
	FindById(context.Context, int64) (User, error)
	StoreUser(User) (User, error)
	UpdateUser(context.Context, User) error
	DeleteById(int64) error
}
