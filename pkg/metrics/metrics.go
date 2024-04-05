package metrics

import (
	"sort"
	"time"

	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/metrics/domain"
)

func NewService(repo domain.Repository) Service {
	return Service{repo: repo}
}

type Service struct {
	repo domain.Repository
}

func (svc Service) GetBalance() (int64, error) {
	return svc.repo.GetBalance()
}

func (svc Service) GetTransactionCount() (int64, error) {
	return svc.repo.GetTransactionCount()
}

func (svc Service) GetUserCount() (int, error) {
	return svc.repo.GetUserCount()
}

func (svc Service) GetArticles() ([]adomain.Article, error) {
	return svc.repo.GetArticles()
}

func (svc Service) GetTransactionsPerDay(days int) ([]domain.Day, error) {
	daysAgo := time.Now().UTC().AddDate(0, 0, -days).Truncate(24 * time.Hour)
	perDay := make(map[string]domain.Day, days)
	for i := range days {
		date := daysAgo.Add(time.Duration(i*24) * time.Hour).Format("2006-01-02")
		perDay[date] = domain.Day{Date: date}
	}
	ds, err := svc.repo.GetTransactionsPerDay(daysAgo)
	if err != nil {
		return nil, err
	}

	for _, day := range ds {
		perDay[day.Date] = day
	}

	result := make([]domain.Day, 0, days)
	for _, day := range perDay {
		result = append(result, day)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date > result[j].Date
	})
	return result, nil
}
