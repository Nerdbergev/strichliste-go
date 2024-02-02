package domain

import (
	"context"
	"time"

	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/users/domain"
)

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
