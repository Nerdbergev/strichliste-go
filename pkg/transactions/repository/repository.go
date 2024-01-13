package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/nerdbergev/shoppinglist-go/pkg/database"
	"github.com/nerdbergev/shoppinglist-go/pkg/transactions/domain"
	udomain "github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
)

var (
	ErrTransactionInvalid = errors.New("Amout can't be positive when sending money or buying an article")
	ErrUserNotFound       = errors.New("User not found")
)

func New(db *sql.DB) TransactionRepository {
	return TransactionRepository{db: db}
}

type TransactionRepository struct {
	db *sql.DB
}

func (r TransactionRepository) FindByUserId(userID int64) ([]domain.Transaction, error) {
	query := `SELECT t.id, t.article_id, t.recipient_transaction_id, t.sender_transaction_id,
        t.quantity, t.comment, t.amount, t.deleted, t.created, u.id, u.name, u.email,
        u.balance, u.disabled, u.created, u.updated
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
		var t Transaction
		var u User
		err := rows.Scan(&t.ID, &t.ArticleID, &t.RecipientID, &t.SenderID, &t.Quantity, &t.Comment,
			&t.Amount, &t.IsDeleted, &t.Created, &u.ID, &u.Name, &u.Email, &u.Balance, &u.IsDisabled,
			&u.Created, &u.Updated)
		if err != nil {
			return nil, err
		}
		t.User = u
		transactions = append(transactions, mapTransactionToDomain(t))
	}

	return transactions, nil
}

func (r TransactionRepository) Transaction(ctx context.Context, f func(ctx context.Context) error) error {
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

func (r TransactionRepository) ProcessTransaction(userID, amount int64, comment string, quantity, articleID, recipientID *int64) (domain.Transaction, error) {
	if (recipientID != nil || articleID != nil) && amount > 0 {
		return domain.Transaction{}, ErrTransactionInvalid
	}

	tx, err := r.db.Begin()
	if err != nil {
		return domain.Transaction{}, err
	}
	defer tx.Rollback()

	user, err := r.findUserById(tx, userID)
	if err != nil {
		return domain.Transaction{}, err
	}

	user.Balance += amount
	_ = r.persistUserBalance(tx, user)
	t := Transaction{
		Created: time.Now(),
		User:    user,
	}

	switch {
	case articleID != nil:
		r.processPurchase(tx, t, articleID, quantity)
	case recipientID != nil:
		// r.processTransfer(tx)
	}

	query := `INSERT INTO transactions (amount, user_id, created, deleted) VALUES ($1, $2, $3, false)`
	res, err := tx.Exec(query, amount, userID, t.Created)
	if err != nil {
		return domain.Transaction{}, err
	}
	t.ID, err = res.LastInsertId()
	if err != nil {
		return domain.Transaction{}, err
	}
	t.Amount = amount
	_ = tx.Commit()
	return mapTransactionToDomain(t), nil
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

func (r TransactionRepository) processPurchase(tx *sql.Tx, t Transaction, articleID, quantity *int64) error {
	row := tx.QueryRow("SELECT * FROM article WHERE id = ?", articleID)
	var a Article
	err := row.Scan(a.ID, a.PrecursorID, a.Name, a.Barcode, a.Amount, a.IsActive, a.Created, a.UsageCount)
	if err != nil {
		return err
	}

	if !a.IsActive {
		return errors.New("Article Inactive")
	}

	if quantity == nil {
		quantity = new(int64)
		*quantity = 1
	}
	t.Quantity.Int64 = *quantity
	t.Amount = a.Amount * *quantity * -1
	t.ArticleID.Int64 = a.ID

	return nil
}

func (r TransactionRepository) StoreTransaction(t domain.Transaction) (domain.Transaction, error) {
	return domain.Transaction{}, nil
}

func (r TransactionRepository) GetAll() ([]domain.Transaction, error) {
	return nil, nil
}

func (r TransactionRepository) FindByUserIdAndTransactionId(uid, tid int64) (domain.Transaction, error) {
	return domain.Transaction{}, nil
}
func (r TransactionRepository) DeleteByUserIdAndTransactionId(uid, tid int64) (domain.Transaction, error) {
	return domain.Transaction{}, nil
}

func (r TransactionRepository) findUserById(tx *sql.Tx, userID int64) (User, error) {
	row := tx.QueryRow("SELECT * FROM user WHERE id = ?", userID)
	return processUserRow(row)
}

func (r TransactionRepository) persistUserBalance(tx *sql.Tx, user User) error {
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
	ArticleID   sql.NullInt64
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
