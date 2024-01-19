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
	query := `SELECT t.id, t.recipient_transaction_id, t.sender_transaction_id,
        t.quantity, t.comment, t.amount, t.deleted, t.created, u.id, u.name, u.email,
        u.balance, u.disabled, u.created, u.updated, a.id, a.name, a.barcode, a.amount,
        a.active, a.created, a.usage_count
    FROM transactions AS t
    JOIN user AS u ON t.user_id = u.id
    JOIN article as a ON t.article_id = a.id
    WHERE t.user_id = ?`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var t Transaction
		var u User
		var a Article
		err := rows.Scan(&t.ID, &t.RecipientID, &t.SenderID, &t.Quantity, &t.Comment,
			&t.Amount, &t.IsDeleted, &t.Created, &u.ID, &u.Name, &u.Email, &u.Balance, &u.IsDisabled,
			&u.Created, &u.Updated, &a.ID, &a.Name, &a.Barcode, &a.Amount, &a.IsActive, &a.Created,
			&a.UsageCount)
		if err != nil {
			return nil, err
		}
		t.User = u
		t.Article = a
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
	query := `INSERT INTO transactions (amount, user_id, created, article_id, deleted) VALUES ($1, $2, $3, $4, false)`
	var articleID *int64
	if t.Article != nil {
		articleID = &t.Article.ID
	}
	res, err := r.getDB(ctx).Exec(query, t.Amount, t.User.ID, t.Created, articleID)
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

func (r Repository) findUserById(tx *sql.Tx, userID int64) (User, error) {
	row := tx.QueryRow("SELECT * FROM user WHERE id = ?", userID)
	return processUserRow(row)
}

func (r Repository) persistUserBalance(tx *sql.Tx, user User) error {
	query := `UPDATE user SET balance = ? WHERE id = ?`
	_, err := tx.Exec(query, user.Balance, user.ID)
	return err
}

func processUserRow(r *sql.Row) (User, error) {
	var user User
	err := r.Scan(&user.ID, &user.Name, &user.Email, &user.Balance, &user.IsDisabled, &user.Created,
		&user.Updated)
	return user, err
}

type Transaction struct {
	ID          int64
	User        User
	Article     Article
	RecipientID sql.NullInt64
	SenderID    sql.NullInt64
	Quantity    sql.NullInt64
	Comment     sql.NullString
	Amount      int64
	IsDeleted   bool
	Created     time.Time
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
		Article:   mapArticleToDomain(t.Article),
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
