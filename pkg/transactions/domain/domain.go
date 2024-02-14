package domain

import (
	"context"
	"fmt"
	"time"

	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/users/domain"
)

type TransactionInvalidError struct {
	Msg string
}

func (e TransactionInvalidError) Error() string {
	return e.Msg
}

type TransactionNotFoundError struct {
	ID int64
}

func (e TransactionNotFoundError) Error() string {
	return fmt.Sprintf("Transaction '%d' not found", e.ID)
}

type TransactionNotDeletableError struct {
	ID int64
}

func (e TransactionNotDeletableError) Error() string {
	// Keep the typo for compatibility
	return fmt.Sprintf("Transaction '%d' is not deleteable", e.ID)
}

type TransactionBoundaryError struct {
	Amount    int64
	Boundrary int64
}

func (e TransactionBoundaryError) Error() string {
	if e.Amount > e.Boundrary {
		return fmt.Sprintf("Transaction amount '%d' exceeds upper transaction boundary '%d'", e.Amount, e.Boundrary)
	} else {
		return fmt.Sprintf("Transaction amount '%d' is below lower transaction boundary '%d'", e.Amount, e.Boundrary)
	}
}

type AccountBalanceBoundaryException struct {
	UID      int64
	Amount   int64
	Boundary int64
}

func (e AccountBalanceBoundaryException) Error() string {
	if e.Amount > e.Boundary {
		return fmt.Sprintf("Transaction amount '%d' exceeds upper account balance boundary '%d' for user '%d'", e.Amount, e.Boundary, e.UID)
	} else {
		return fmt.Sprintf("Transaction amount '%d' is below lower account balance boundary '%d' for user '%d'", e.Amount, e.Boundary, e.UID)
	}
}

type Transaction struct {
	ID                   int64
	User                 domain.User
	Article              *adomain.Article
	RecipientTransaction *Transaction
	SenderTransaction    *Transaction
	Quantity             *int64
	Comment              *string
	Amount               int64
	IsDeleted            bool
	IsDeletable          bool
	Created              time.Time
}

type TransactionRepository interface {
	Transactional(context.Context, func(context.Context) error) error
	GetAll() ([]Transaction, error)
	StoreTransaction(context.Context, Transaction) (Transaction, error)
	FindByUserId(int64) ([]Transaction, error)
	FindById(context.Context, int64) (Transaction, error)
	DeleteById(context.Context, int64) error
	UpdateSenderTransaction(context.Context, Transaction) error
	MarkDeleted(context.Context, int64) error
}
