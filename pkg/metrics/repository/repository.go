package repository

import (
	"database/sql"
	"time"

	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/metrics/domain"
)

func New(db *sql.DB) Repository {
	return Repository{db: db}
}

type Repository struct {
	db *sql.DB
}

func (repo Repository) GetBalance() (int64, error) {
	row := repo.db.QueryRow("SELECT SUM(balance) from user where disabled = false")
	var balance int64
	err := row.Scan(&balance)
	return balance, err
}

func (repo Repository) GetTransactionCount() (int64, error) {
	row := repo.db.QueryRow("SELECT COUNT(*) from transactions")
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (repo Repository) GetUserCount() (int, error) {
	row := repo.db.QueryRow("SELECT COUNT(*) from user")
	var count int
	err := row.Scan(&count)
	return count, err
}

func (repo Repository) GetArticles() ([]adomain.Article, error) {
	rows, err := repo.db.Query("SELECT * from article WHERE active = true order by usage_count desc")
	if err != nil {
		return nil, err
	}
	var articles []adomain.Article
	for rows.Next() {
		var a Article
		err := rows.Scan(&a.ID, &a.PrecursorID, &a.Name, &a.Barcode, &a.Amount, &a.IsActive, &a.Created,
			&a.UsageCount)
		if err != nil {
			return nil, err
		}
		articles = append(articles, *mapArticleToDomain(a))
	}

	return articles, nil
}

func (repo Repository) GetTransactionsPerDay(start time.Time) ([]domain.Day, error) {
	rows, err := repo.db.Query(`SELECT
        DATE(created) as createDate,
        COUNT(id),
        SUM((CASE WHEN amount >= 0 THEN 1 ELSE 0 END)),
        SUM((CASE WHEN amount < 0 THEN 1 ELSE 0 END)),
        COUNT(DISTINCT user_id),
        SUM(amount),
        SUM((CASE WHEN amount >= 0 THEN amount ELSE 0 END)),
        SUM((CASE WHEN amount < 0 THEN amount ELSE 0 END))
        FROM transactions WHERE created >= ? GROUP BY createDate ORDER BY createDate`, start)
	if err != nil {
		return nil, err
	}
	var days []domain.Day
	for rows.Next() {
		var (
			day           domain.Day
			chargedCount  int
			spentCount    int
			chargedAmount int
			spentAmount   int
		)
		err := rows.Scan(&day.Date, &day.TransactionCount, &chargedCount, &spentCount, &day.DistinctUserCount, &day.Balance, &chargedAmount, &spentAmount)
		if err != nil {
			return nil, err
		}

		if chargedCount != 0 {
			day.Charged = &domain.Turnover{
				Amount:            chargedAmount,
				TransactionsCount: chargedCount,
			}
		}

		if spentCount != 0 {
			day.Spent = &domain.Turnover{
				Amount:            spentAmount,
				TransactionsCount: spentCount,
			}
		}
		days = append(days, day)
	}
	return days, nil
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
