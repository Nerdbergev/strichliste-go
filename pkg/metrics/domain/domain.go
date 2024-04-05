package domain

import (
	"time"

	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
)

type Turnover struct {
	Amount            int
	TransactionsCount int
}

type Day struct {
	Date              string
	TransactionCount  int
	DistinctUserCount int
	Balance           int64
	Charged           *Turnover
	Spent             *Turnover
}

type Metric struct {
	Balance          int64
	TransactionCount int
	UserCount        int
	Articles         []adomain.Article
	Days             []Day
}

type Repository interface {
	GetBalance() (int64, error)
	GetTransactionCount() (int64, error)
	GetUserCount() (int, error)
	GetArticles() ([]adomain.Article, error)
	GetTransactionsPerDay(time.Time) ([]Day, error)
}
