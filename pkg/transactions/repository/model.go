package repository

import (
	"database/sql"
	"time"

	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/transactions/domain"
	udomain "github.com/nerdbergev/strichliste-go/pkg/users/domain"
)

type Transaction struct {
	ID                   int64
	User                 User
	Article              *Article
	RecipientTransaction *Transaction
	SenderTransaction    *Transaction
	Quantity             sql.NullInt64
	Comment              sql.NullString
	Amount               int64
	IsDeleted            bool
	Created              time.Time
}

type User struct {
	ID         int64
	Name       string
	Email      sql.NullString
	Balance    int64
	IsDisabled bool
	Created    time.Time
	Updated    sql.NullTime
}

type Article struct {
	ID          int64
	PrecursorID sql.NullInt64
	Name        string
	Barcode     sql.NullString
	Amount      int64
	IsActive    bool
	Created     time.Time
	UsageCount  int64
}

func mapTransactionToDomain(t Transaction) domain.Transaction {
	dt := domain.Transaction{
		ID:        t.ID,
		User:      mapUserToDomain(t.User),
		Amount:    t.Amount,
		IsDeleted: t.IsDeleted,
		Created:   t.Created,
	}

	if t.Quantity.Valid {
		dt.Quantity = new(int64)
		dt.Quantity = &t.Quantity.Int64
	}

	if t.Comment.Valid {
		dt.Comment = new(string)
		dt.Comment = &t.Comment.String
	}

	if t.Article != nil {
		dt.Article = mapArticleToDomain(*t.Article)
	}

	if t.SenderTransaction != nil {
		dt.SenderTransaction = new(domain.Transaction)
		*dt.SenderTransaction = mapTransactionToDomain(*t.SenderTransaction)
	}

	if t.RecipientTransaction != nil {
		dt.RecipientTransaction = new(domain.Transaction)
		*dt.RecipientTransaction = mapTransactionToDomain(*t.RecipientTransaction)
	}
	return dt
}

func mapUserToDomain(u User) udomain.User {
	du := udomain.User{
		ID:         u.ID,
		Name:       u.Name,
		Balance:    u.Balance,
		IsDisabled: u.IsDisabled,
		Created:    u.Created,
	}
	if u.Email.Valid {
		du.Email = new(string)
		*du.Email = u.Email.String
	}
	if u.Updated.Valid {
		du.Updated = new(time.Time)
		*du.Updated = u.Updated.Time
	}
	return du
}

func mapArticleToDomain(a Article) *adomain.Article {
	da := &adomain.Article{
		ID:         a.ID,
		Name:       a.Name,
		Amount:     a.Amount,
		IsActive:   a.IsActive,
		Created:    a.Created,
		UsageCount: a.UsageCount,
	}

	if a.Barcode.Valid {
		da.Barcode = new(string)
		*da.Barcode = a.Barcode.String
	}

	return da
}
