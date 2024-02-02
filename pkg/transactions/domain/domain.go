package domain

import (
	"context"
	"time"

	adomain "github.com/nerdbergev/shoppinglist-go/pkg/articles/domain"
	"github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
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
	Transaction(context.Context, func(context.Context) error) error
	GetAll() ([]Transaction, error)
	StoreTransaction(context.Context, Transaction) (Transaction, error)
	FindByUserId(int64) ([]Transaction, error)
	FindByUserIdAndTransactionId(uid, tid int64) (Transaction, error)
	DeleteByUserIdAndTransactionId(uid, tid int64) (Transaction, error)
	UpdateSenderTransaction(context.Context, Transaction) error
}
