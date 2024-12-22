package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

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

type Barcode struct {
	ID        int64
	Barcode   string
	ArticleID int64
	Created   time.Time
}

type BarcodeRepository interface {
	GetAll() ([]Barcode, error)
	FindByArticleID(int64) ([]Barcode, error)
	FindByID(int64) (Barcode, error)
	FindByBarcode(string) (Barcode, error)
	ReassignBarcode(context.Context, int64, int64) error
	StoreBarcode(Barcode) (Barcode, error)
	DeleteByID(int64) error
}
