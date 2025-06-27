package repository

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/database"
	"github.com/pkg/errors"
)

var (
	ErrBarcodeNotFound = errors.New("Barcode not found")
)

func NewBarcodeRepository(db *sql.DB) BarcodeRepository {
	return BarcodeRepository{db: db}
}

type BarcodeRepository struct {
	db *sql.DB
}

func (r BarcodeRepository) GetAll() ([]domain.Barcode, error) {
	rows, err := r.db.Query("SELECT * FROM barcode")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return processBarcodeRows(rows)
}

func (r BarcodeRepository) FindByArticleID(aid domain.ArticleID) ([]domain.Barcode, error) {
	rows, err := r.db.Query("SELECT * FROM barcode where article_id=?", aid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return processBarcodeRows(rows)
}

func (r BarcodeRepository) FindByID(bid domain.BarcodeID) (domain.Barcode, error) {
	row := r.db.QueryRow("SELECT * FROM barcode WHERE id = ?", bid)
	barcode, err := processBarcodeRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Barcode{}, errors.Wrap(
				domain.BarcodeNotFoundError{Identifier: strconv.FormatInt(int64(bid), 10)},
				err.Error(),
			)
		}
		return domain.Barcode{}, errors.Wrap(domain.ErrPersistanceError, err.Error())
	}
	return barcode, nil
}

func (r BarcodeRepository) FindByBarcode(barcode string) (domain.Barcode, error) {
	row := r.db.QueryRow("SELECT * FROM barcode b WHERE b.barcode = ?", barcode)
	b, err := processBarcodeRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Barcode{}, errors.Wrap(
				domain.BarcodeNotFoundError{Identifier: barcode},
				err.Error(),
			)
		}
		return domain.Barcode{}, errors.Wrap(domain.ErrPersistanceError, err.Error())
	}
	return b, nil
}

func (r BarcodeRepository) ReassignBarcode(ctx context.Context, bid domain.BarcodeID, aid domain.ArticleID) error {
	_, err := r.getDB(ctx).Exec("UPDATE barcode SET article_id=? WHERE id =?", aid, bid)
	return err
}

func (r BarcodeRepository) Store(b domain.Barcode) (domain.Barcode, error) {
	res, err := r.db.Exec("INSERT INTO barcode (article_id, barcode, created) VALUES (?, ?, ?)", b.ArticleID,
		b.Barcode, b.Created)
	if err != nil {
		return domain.Barcode{}, err
	}
	insertID, err := res.LastInsertId()
	if err != nil {
		return domain.Barcode{}, err
	}
	b.ID = domain.BarcodeID(insertID)

	return b, nil
}

func (r BarcodeRepository) DeleteByID(bid domain.BarcodeID) error {
	_, err := r.db.Exec("DELETE FROM barcode WHERE id = ?", bid)
	return err
}

func (r BarcodeRepository) getDB(ctx context.Context) database.DB {
	if db, ok := database.FromContext(ctx); ok {
		return db
	}
	return r.db
}

type Barcode struct {
	ID        int64
	Barcode   string
	ArticleID int64
	Created   time.Time
}

func processBarcodeRows(r *sql.Rows) ([]domain.Barcode, error) {
	var barcodes []domain.Barcode
	for r.Next() {
		var b Barcode
		err := r.Scan(&b.ID, &b.ArticleID, &b.Barcode, &b.Created)
		if err != nil {
			return nil, err
		}
		barcodes = append(barcodes, mapBarcodeToDomain(b))
	}
	return barcodes, nil
}

func processBarcodeRow(r *sql.Row) (domain.Barcode, error) {
	var b Barcode
	err := r.Scan(&b.ID, &b.ArticleID, &b.Barcode, &b.Created)
	if err != nil {
		return domain.Barcode{}, err
	}
	return mapBarcodeToDomain(b), nil
}

func mapBarcodeToDomain(b Barcode) domain.Barcode {
	db := domain.Barcode{
		ID:        domain.BarcodeID(b.ID),
		Barcode:   b.Barcode,
		ArticleID: domain.ArticleID(b.ArticleID),
		Created:   b.Created,
	}
	return db
}
