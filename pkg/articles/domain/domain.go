package domain

import (
	"context"
	"time"
)

type Article struct {
	ID         int64
	Name       string
	Barcode    *string
	Amount     int64
	IsActive   bool
	Created    time.Time
	UsageCount int64
}

func (a *Article) IncrementUsageCount() {
	a.UsageCount += 1
}

func (a *Article) DecrementUsageCount() {
	a.UsageCount -= 1
}

type ArticleRepository interface {
	GetAll(bool) ([]Article, error)
	FindById(context.Context, int64) (Article, error)
	FindByBarcode(string) (Article, error)
	StoreArticle(Article) (Article, error)
	UpdateArticle(context.Context, Article) error
	DeleteById(int64) error
}
