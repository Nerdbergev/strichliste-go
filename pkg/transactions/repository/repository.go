package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/nerdbergev/strichliste-go/pkg/database"
	"github.com/nerdbergev/strichliste-go/pkg/transactions/domain"
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
        SELECT t.id, t.article_id, t.recipient_transaction_id, t.sender_transaction_id, t.quantity,
        t.comment, t.amount, t.deleted, t.created, u.id, u.name, u.email, u.balance, u.disabled,
        u.created, u.updated
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
			&t.Amount, &t.IsDeleted, &t.Created, &u.ID, &u.Name, &u.Email, &u.Balance,
			&u.IsDisabled, &u.Created, &u.Updated)
		if err != nil {
			return nil, err
		}

		if articleID.Valid {
			article, err := r.findArticleById(context.Background(), articleID.Int64)
			if err != nil {
				return nil, err
			}
			t.Article = new(Article)
			*t.Article = article
		}

		if recipientID.Valid {
			recipientTransaction, err := r.findByIdWithoutNestedTransactions(context.Background(), recipientID.Int64)
			if err != nil {
				return nil, err
			}
			t.RecipientTransaction = new(Transaction)
			*t.RecipientTransaction = recipientTransaction
		}

		if senderID.Valid {
			senderTransaction, err := r.findByIdWithoutNestedTransactions(context.Background(), senderID.Int64)
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

func (r Repository) Transactional(ctx context.Context, f func(ctx context.Context) error) error {
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

func (r Repository) StoreTransaction(ctx context.Context,
	t domain.Transaction) (domain.Transaction, error) {
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
	res, err := r.getDB(ctx).
		Exec(query, t.Amount, t.User.ID, t.Created, articleID, recipientTransactionID)
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

func (r Repository) FindById(ctx context.Context, tid int64) (domain.Transaction, error) {
	query := `
        SELECT t.id, t.article_id, t.recipient_transaction_id, t.sender_transaction_id, t.quantity,
        t.comment, t.amount, t.deleted, t.created, u.id, u.name, u.email, u.balance, u.disabled,
        u.created, u.updated
        FROM transactions AS t
	    JOIN user AS u ON t.user_id = u.id
	    WHERE t.id = ?`
	row := r.getDB(ctx).QueryRow(query, tid)

	var (
		t           Transaction
		u           User
		articleID   sql.NullInt64
		recipientID sql.NullInt64
		senderID    sql.NullInt64
	)
	err := row.Scan(&t.ID, &articleID, &recipientID, &senderID, &t.Quantity, &t.Comment,
		&t.Amount, &t.IsDeleted, &t.Created, &u.ID, &u.Name, &u.Email, &u.Balance,
		&u.IsDisabled, &u.Created, &u.Updated)
	if err != nil {
		return domain.Transaction{}, err
	}

	if articleID.Valid {
		article, err := r.findArticleById(ctx, articleID.Int64)
		if err != nil {
			return domain.Transaction{}, err
		}
		t.Article = new(Article)
		*t.Article = article
	}

	if recipientID.Valid {
		recipientTransaction, err := r.findByIdWithoutNestedTransactions(ctx, recipientID.Int64)
		if err != nil {
			return domain.Transaction{}, err
		}
		t.RecipientTransaction = new(Transaction)
		*t.RecipientTransaction = recipientTransaction
	}

	if senderID.Valid {
		senderTransaction, err := r.findByIdWithoutNestedTransactions(ctx, senderID.Int64)
		if err != nil {
			return domain.Transaction{}, err
		}
		t.SenderTransaction = new(Transaction)
		*t.SenderTransaction = senderTransaction
	}

	t.User = u

	return mapTransactionToDomain(t), nil

}

func (r Repository) DeleteById(ctx context.Context, tid int64) error {
	query := `DELETE FROM transactions WHERE id = $1`
	_, err := r.getDB(ctx).Exec(query, tid)
	return err
}

func (r Repository) UpdateSenderTransaction(ctx context.Context, t domain.Transaction) error {
	query := `UPDATE transactions SET sender_transaction_id = $1 WHERE id = $2`
	_, err := r.getDB(ctx).Exec(query, t.SenderTransaction.ID, t.ID)
	return err
}

func (r Repository) MarkDeleted(ctx context.Context, tid int64) error {
	query := `UPDATE transactions SET deleted = true WHERE id = $1`
	_, err := r.getDB(ctx).Exec(query, tid)
	return err
}

func (r Repository) findArticleById(ctx context.Context, aid int64) (Article, error) {
	row := r.getDB(ctx).QueryRow("SELECT * FROM article WHERE id = ?", aid)
	var a Article
	err := row.Scan(&a.ID, &a.PrecursorID, &a.Name, &a.Barcode, &a.Amount, &a.IsActive, &a.Created,
		&a.UsageCount)
	return a, err
}

func (r Repository) findByIdWithoutNestedTransactions(ctx context.Context, id int64) (Transaction, error) {
	query := `
        SELECT t.id, t.article_id, t.quantity, t.comment, t.amount, t.deleted, t.created,
               u.id, u.name, u.email, u.balance, u.disabled, u.created, u.updated
        FROM transactions AS t
	    JOIN user AS u ON t.user_id = u.id
	    WHERE t.id = ?`
	row := r.getDB(ctx).QueryRow(query, id)

	var (
		transaction Transaction
		user        User
		articleID   sql.NullInt64
	)
	err := row.Scan(&transaction.ID, &articleID, &transaction.Quantity, &transaction.Comment,
		&transaction.Amount, &transaction.IsDeleted, &transaction.Created, &user.ID, &user.Name,
		&user.Email, &user.Balance, &user.IsDisabled, &user.Created, &user.Updated)

	if err != nil {
		return Transaction{}, err
	}

	transaction.User = user

	if articleID.Valid {
		article, err := r.findArticleById(ctx, articleID.Int64)
		if err != nil {
			return Transaction{}, err
		}
		transaction.Article = new(Article)
		*transaction.Article = article
	}

	return transaction, nil
}

func (r Repository) getDB(ctx context.Context) database.DB {
	if db, ok := database.FromContext(ctx); ok {
		return db
	}
	return r.db
}
