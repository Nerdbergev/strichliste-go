package model

import (
	"database/sql"
	"time"
)

func NewUserRepository(db *sql.DB) TransactionRepository {
	return TransactionRepository{db: db}
}

type TransactionRepository struct {
	db *sql.DB
}

func (r TransactionRepository) GetFromUser(userID int64) ([]Transaction, error) {
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

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		var u User
		err := rows.Scan(&t.ID, &t.ArticleID, &t.RecipientID, &t.SenderID, &t.Quantity, &t.Comment,
			&t.Amount, &t.IsDeleted, &t.Created, &u.ID, &u.Name, &u.Email, &u.Balance, &u.Disabled,
			&u.Created, &u.Updated)
		if err != nil {
			return nil, err
		}
		t.User = u
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func (r TransactionRepository) Deposit(userID, amount int64) (Transaction, error) {
	t := Transaction{}
	t.Created = time.Now()
	query := `INSERT INTO transactions (amount, user_id, created, deleted) VALUES ($1, $2, $3, false)`
	res, err := r.db.Exec(query, amount, userID, t.Created)
	if err != nil {
		return Transaction{}, err
	}
	t.ID, err = res.LastInsertId()
	if err != nil {
		return Transaction{}, err
	}

	return t, nil
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
	ID       int64
	Name     string
	Email    string
	Balance  int
	Disabled bool
	Created  time.Time
	Updated  *time.Time
}
