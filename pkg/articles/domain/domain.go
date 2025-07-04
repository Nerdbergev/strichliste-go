package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type (
	ArticleID int64
	BarcodeID int64
	TagID     int64
)

type ArticleNotFoundError struct {
	Identifier string
}

func (err ArticleNotFoundError) Error() string {
	return fmt.Sprintf("Article '%s' not found", err.Identifier)
}

type ArticleInactiveError struct {
	Name string
	Id   ArticleID
}

func (err ArticleInactiveError) Error() string {
	return fmt.Sprintf("Article '%s' (%d) is inactive", err.Name, err.Id)
}

type ArticleBarcodeAlreadyExistsError struct {
	Id      ArticleID
	Barcode string
}

func (err ArticleBarcodeAlreadyExistsError) Error() string {
	return fmt.Sprintf("Active article (%d) with barcode '%s' already exists.", err.Id, err.Barcode)
}

type Tag struct {
	ID         TagID
	Tag        string
	Created    time.Time
	UsageCount int64
}

type Barcode struct {
	ID        BarcodeID
	ArticleID ArticleID
	Barcode   string
	Created   time.Time
}

type Article struct {
	ID         ArticleID
	Name       string
	Barcodes   []Barcode
	Tags       []Tag
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

func (a Article) IsActivatable() bool {
	if a.IsActive {
		return false
	}

	if a.Precursor != nil {
		return false
	}

	return true
}

type ArticleRepository interface {
	GetAll(bool, bool, string, *bool) ([]Article, error)
	CountActive() (int, error)
	FindById(context.Context, ArticleID) (Article, error)
	FindActiveByBarcode(string) (Article, error)
	Store(context.Context, Article) (Article, error)
	Update(context.Context, Article) error
	DeleteById(ArticleID) error
	Transactional(context.Context, func(context.Context) error) error
}

type BarcodeRepository interface {
	GetAll() ([]Barcode, error)
	FindByArticleID(ArticleID) ([]Barcode, error)
	FindByID(BarcodeID) (Barcode, error)
	FindByBarcode(string) (Barcode, error)
	ReassignBarcode(context.Context, BarcodeID, ArticleID) error
	Store(Barcode) (Barcode, error)
	DeleteByID(BarcodeID) error
}

type TagRepository interface {
	GetAll() ([]Tag, error)
	GetArticleTags(ArticleID) ([]Tag, error)
	FindByID(TagID) (Tag, error)
	FindByTag(string) (Tag, error)
	CheckArticleHasTag(ArticleID, TagID) bool
	Transactional(context.Context, func(context.Context) error) error
	CreateTag(context.Context, string, time.Time) (Tag, error)
	AddArticleTag(context.Context, ArticleID, TagID, time.Time) error
	DeleteArticleTag(context.Context, ArticleID, TagID) error
	DeleteTag(context.Context, TagID) error
}

var (
	ErrPersistanceError = errors.New("persistance error")
)

type BarcodeNotFoundError struct {
	Identifier string
	Cause      error
}

func (e BarcodeNotFoundError) Error() string {
	return fmt.Sprintf("Barcode ID '%s' not found", e.Identifier)
}
