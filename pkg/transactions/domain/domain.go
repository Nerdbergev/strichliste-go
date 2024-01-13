package domain

import (
	"time"

	adomain "github.com/nerdbergev/shoppinglist-go/pkg/articles/domain"
	"github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
)

type Transaction struct {
	ID        int64
	User      domain.User
	Article   *adomain.Article
	Recipient *domain.User
	Sender    *domain.User
	Quantity  *int64
	Comment   *string
	Amount    int64
	IsDeleted bool
	Created   time.Time
}

type TransactionRepository interface {
	GetAll() ([]Transaction, error)
	StoreTransaction(Transaction) (Transaction, error)
	FindByUserId(int64) ([]Transaction, error)
	FindByUserIdAndTransactionId(uid, tid int64) (Transaction, error)
	DeleteByUserIdAndTransactionId(uid, tid int64) (Transaction, error)
}
