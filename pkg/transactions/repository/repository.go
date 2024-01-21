package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	adomain "github.com/nerdbergev/shoppinglist-go/pkg/articles/domain"
	"github.com/nerdbergev/shoppinglist-go/pkg/database"
	"github.com/nerdbergev/shoppinglist-go/pkg/transactions/domain"
	udomain "github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
)

var (
	ErrTransactionInvalid = errors.New("Amout can't be positive when sending money or buying an article")
	ErrUserNotFound       = errors.New("User not found")
)

func New(db *sql.DB) Repository {
	return Repository{db: db}
}

type Repository struct {
	db *sql.DB
}

func (r Repository) FindByUserId(userID int64) ([]domain.Transaction, error) {
	query := `
        SELECT t.id, t.article_id, t.recipient_transaction_id, t.sender_transaction_id, t.quantity, t.comment,
        t.amount, t.deleted, t.created, u.id, u.name, u.email, u.balance, u.disabled, u.created,
        u.updated
        FROM transactions AS t
	    JOIN user AS u ON t.user_id = u.id
	    WHERE t.user_id = ?`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var (
			t           Transaction
			u           User
			articleID   sql.NullInt64
			recipientID sql.NullInt64
			senderID    sql.NullInt64
		)
		err := rows.Scan(&t.ID, &articleID, &recipientID, &senderID, &t.Quantity, &t.Comment,
			&t.Amount, &t.IsDeleted, &t.Created, &u.ID, &u.Name, &u.Email, &u.Balance, &u.IsDisabled,
			&u.Created, &u.Updated)
		if err != nil {
			return nil, err
		}
		if articleID.Valid {
			article, err := r.findArticleById(articleID.Int64)
			if err != nil {
				return nil, err
			}
			t.Article = new(Article)
			*t.Article = article
		}
		if recipientID.Valid {
			recipientTransaction, err := r.findById(recipientID.Int64)
			if err != nil {
				return nil, err
			}
			t.RecipientTransaction = new(Transaction)
			*t.RecipientTransaction = recipientTransaction
		}

		if senderID.Valid {
			senderTransaction, err := r.findById(senderID.Int64)
			if err != nil {
				return nil, err
			}
			t.SenderTransaction = new(Transaction)
			*t.SenderTransaction = senderTransaction
		}

		t.User = u
		transactions = append(transactions, mapTransactionToDomain(t))
	}

	return transactions, nil
}

func (r Repository) Transaction(ctx context.Context, f func(ctx context.Context) error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = f(database.AddToContext(ctx, tx))
	if err != nil {
		return err
	}

	return tx.Commit()
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

func (r Repository) StoreTransaction(ctx context.Context, t domain.Transaction) (domain.Transaction, error) {
	query := `INSERT INTO transactions
                (amount, user_id, created, article_id, deleted, recipient_transaction_id)
                VALUES ($1, $2, $3, $4, false, $5)`
	var (
		articleID              *int64
		recipientTransactionID *int64
	)
	if t.Article != nil {
		articleID = &t.Article.ID
	}
	if t.RecipientTransaction != nil {
		recipientTransactionID = &t.RecipientTransaction.ID
	}
	res, err := r.getDB(ctx).Exec(query, t.Amount, t.User.ID, t.Created, articleID, recipientTransactionID)
	if err != nil {
		return domain.Transaction{}, err
	}
	t.ID, err = res.LastInsertId()
	if err != nil {
		return domain.Transaction{}, err
	}
	return t, nil
}

func (r Repository) GetAll() ([]domain.Transaction, error) {
	return nil, nil
}

func (r Repository) FindByUserIdAndTransactionId(uid, tid int64) (domain.Transaction, error) {
	return domain.Transaction{}, nil
}
func (r Repository) DeleteByUserIdAndTransactionId(uid, tid int64) (domain.Transaction, error) {
	return domain.Transaction{}, nil
}

func (r Repository) UpdateSenderTransaction(ctx context.Context, t domain.Transaction) error {
	query := `UPDATE transactions SET sender_transaction_id = $1 WHERE id = $2`
	_, err := r.getDB(ctx).Exec(query, t.SenderTransaction.ID, t.ID)
	return err
}

func (r Repository) findArticleById(aid int64) (Article, error) {
	row := r.db.QueryRow("SELECT * FROM article WHERE id = ?", aid)
	var a Article
	err := row.Scan(&a.ID, &a.PrecursorID, &a.Name, &a.Barcode, &a.Amount, &a.IsActive, &a.Created, &a.UsageCount)
	return a, err
}

func (r Repository) findById(id int64) (Transaction, error) {
	query := `
        SELECT t.id, t.article_id, t.quantity, t.comment, t.amount, t.deleted, t.created,
               u.id, u.name, u.email, u.balance, u.disabled, u.created, u.updated
        FROM transactions AS t
	    JOIN user AS u ON t.user_id = u.id
	    WHERE t.id = ?`
	row := r.db.QueryRow(query, id)

	var (
		t         Transaction
		u         User
		articleID sql.NullInt64
	)
	err := row.Scan(&t.ID, &articleID, &t.Quantity, &t.Comment,
		&t.Amount, &t.IsDeleted, &t.Created, &u.ID, &u.Name, &u.Email, &u.Balance, &u.IsDisabled,
		&u.Created, &u.Updated)

	if err != nil {
		return Transaction{}, err
	}

	t.User = u

	if articleID.Valid {
		article, err := r.findArticleById(articleID.Int64)
		if err != nil {
			return Transaction{}, err
		}
		t.Article = new(Article)
		*t.Article = article
	}

	return t, nil
}

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

func (r Repository) getDB(ctx context.Context) database.DB {
	if db, ok := database.FromContext(ctx); ok {
		return db
	}
	return r.db
}
