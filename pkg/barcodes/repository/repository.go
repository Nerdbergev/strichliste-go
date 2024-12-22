package repository

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/nerdbergev/strichliste-go/pkg/barcodes/domain"
	"github.com/nerdbergev/strichliste-go/pkg/database"
	"github.com/pkg/errors"
)

var (
	ErrBarcodeNotFound = errors.New("Barcode not found")
)

func New(db *sql.DB) Repository {
	return Repository{db: db}
}

type Repository struct {
	db *sql.DB
}

func (r Repository) GetAll() ([]domain.Barcode, error) {
	rows, err := r.db.Query("SELECT * FROM barcode")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return processRows(rows)
}

func (r Repository) FindByArticleID(aid int64) ([]domain.Barcode, error) {
	rows, err := r.db.Query("SELECT * FROM barcode where article_id=?", aid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return processRows(rows)
}

func (r Repository) FindByID(bid int64) (domain.Barcode, error) {
	row := r.db.QueryRow("SELECT * FROM barcode WHERE id = ?", bid)
	barcode, err := processRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Barcode{}, errors.Wrap(
				domain.BarcodeNotFoundError{Identifier: strconv.FormatInt(bid, 10)},
				err.Error(),
			)
		}
		return domain.Barcode{}, errors.Wrap(domain.ErrPersistanceError, err.Error())
	}
	return barcode, nil
}

func (r Repository) FindByBarcode(barcode string) (domain.Barcode, error) {
	row := r.db.QueryRow("SELECT * FROM barcode b WHERE b.barcode = ?", barcode)
	b, err := processRow(row)
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

func (r Repository) ReassignBarcode(ctx context.Context, bid, aid int64) error {
	_, err := r.getDB(ctx).Exec("UPDATE barcode SET article_id=? WHERE id =?", aid, bid)
	return err
}

func (r Repository) StoreBarcode(b domain.Barcode) (domain.Barcode, error) {
	res, err := r.db.Exec("INSERT INTO barcode (article_id, barcode, created) VALUES (?, ?, ?)", b.ArticleID,
		b.Barcode, b.Created)
	if err != nil {
		return domain.Barcode{}, err
	}
	b.ID, err = res.LastInsertId()
	if err != nil {
		return domain.Barcode{}, err
	}

	return b, nil
}

func (r Repository) DeleteByID(bid int64) error {
	_, err := r.db.Exec("DELETE FROM barcode WHERE id = ?", bid)
	return err
}

func (r Repository) getDB(ctx context.Context) database.DB {
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

func processRows(r *sql.Rows) ([]domain.Barcode, error) {
	var barcodes []domain.Barcode
	for r.Next() {
		var b Barcode
		err := r.Scan(&b.ID, &b.Barcode, &b.ArticleID, &b.Created)
		if err != nil {
			return nil, err
		}
		barcodes = append(barcodes, mapToDomain(b))
	}
	return barcodes, nil
}

func processRow(r *sql.Row) (domain.Barcode, error) {
	var b Barcode
	err := r.Scan(&b.ID, &b.ArticleID, &b.Barcode, &b.Created)
	if err != nil {
		return domain.Barcode{}, err
	}
	return mapToDomain(b), nil
}

func mapToDomain(b Barcode) domain.Barcode {
	db := domain.Barcode{
		ID:        b.ID,
		Barcode:   b.Barcode,
		ArticleID: b.ArticleID,
		Created:   b.Created,
	}
	return db
}
