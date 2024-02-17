package domain

import (
	"context"
	"fmt"
	"time"
)

type ArticleNotFoundError struct {
	Identifier string
}

func (err ArticleNotFoundError) Error() string {
	return fmt.Sprintf("Article '%s' not found", err.Identifier)
}

type ArticleInactiveError struct {
	Name string
	Id   int64
}

func (err ArticleInactiveError) Error() string {
	return fmt.Sprintf("Article '%s' (%d) is inactive", err.Name, err.Id)
}

type ArticleBarcodeAlreadyExistsError struct {
	Id      int64
	Barcode string
}

func (err ArticleBarcodeAlreadyExistsError) Error() string {
	return fmt.Sprintf("Active article (%d) with barcode '%s' already exists.", err.Id, err.Barcode)
}

type Article struct {
	ID         int64
	Name       string
	Barcode    *string
	Amount     int64
	IsActive   bool
	Created    time.Time
	UsageCount int64
	Precursor  *Article
}

func (a *Article) IncrementUsageCount() {
	a.UsageCount += 1
}

func (a *Article) DecrementUsageCount() {
	a.UsageCount -= 1
}

type ArticleRepository interface {
	GetAll(bool, bool, string, *bool) ([]Article, error)
	FindById(context.Context, int64) (Article, error)
	FindActiveByBarcode(string) (Article, error)
	StoreArticle(context.Context, Article) (Article, error)
	UpdateArticle(context.Context, Article) error
	DeleteById(int64) error
	Transactional(context.Context, func(context.Context) error) error
}
