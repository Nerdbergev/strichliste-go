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
	rows, err := repo.db.Query(`SELECT a1.id, a1.precursor_id, a1.name, a1.amount, a1.active,
        a1.created, a1.usage_count, b.id, b.barcode, b.created
        FROM article a1
        LEFT JOIN barcode b ON (b.article_id = a1.id)
        WHERE active = true order by usage_count desc`)
	if err != nil {
		return nil, err
	}

	articles, err := processRows(rows)
	if err != nil {
		return nil, err
	}

	mapped := make([]adomain.Article, 0, len(articles))
	for _, a := range articles {
		mapped = append(mapped, mapArticleToDomain(a))
	}
	return mapped, nil
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
		err := rows.Scan(&day.Date, &day.TransactionCount, &chargedCount, &spentCount,
			&day.DistinctUserCount, &day.Balance, &chargedAmount, &spentAmount)
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

func processRows(r *sql.Rows) ([]Article, error) {
	articles := make(map[int64]*Article)
	for r.Next() {
		var (
			a Article
			b Barcode
		)

		err := r.Scan(&a.ID, &a.PrecursorID, &a.Name, &a.Amount, &a.IsActive, &a.Created,
			&a.UsageCount, &b.ID, &b.Barcode, &b.Created)
		if err != nil {
			return nil, err
		}
		article, ok := articles[a.ID]
		if !ok {
			article = &a
			articles[a.ID] = article
		}

		if b.ID.Valid {
			article.Barcodes = append(article.Barcodes, b)
		}
	}

	result := make([]Article, 0, len(articles))
	for _, article := range articles {
		result = append(result, *article)
	}
	return result, nil
}

type Barcode struct {
	ID      sql.NullInt64
	Barcode sql.NullString
	Created sql.NullTime
}

type Article struct {
	ID          int64
	PrecursorID sql.NullInt64
	Name        string
	Barcodes    []Barcode
	Amount      int64
	IsActive    bool
	Created     time.Time
	UsageCount  int64
}

func mapArticleToDomain(a Article) adomain.Article {
	da := adomain.Article{
		ID:         adomain.ArticleID(a.ID),
		Name:       a.Name,
		Amount:     a.Amount,
		IsActive:   a.IsActive,
		Created:    a.Created,
		UsageCount: a.UsageCount,
	}
	for _, b := range a.Barcodes {
		da.Barcodes = append(da.Barcodes, adomain.Barcode{
			ID:      adomain.BarcodeID(b.ID.Int64),
			Barcode: b.Barcode.String,
			Created: b.Created.Time,
		})
	}
	return da
}
