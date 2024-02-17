package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/database"
)

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

func New(db *sql.DB) Repository {
	return Repository{db: db}
}

type Repository struct {
	db *sql.DB
}

func (r Repository) GetAll(onlyActive, precursor bool, barcode string, ancestor *bool) ([]domain.Article, error) {
	query := "SELECT a1.id, a1.precursor_id, a1.name, a1.barcode, a1.amount, a1.active, a1.created, a1.usage_count FROM article AS a1"
	whereClauseStarted := false
	if ancestor != nil {
		query += " LEFT JOIN article a2 ON (a2.precursor_id = a1.id)"
		if *ancestor {
			query += " WHERE a2.id IS NOT NULL"
		} else {
			query += " WHERE a2.id IS NULL"
		}
		whereClauseStarted = true
	}

	if onlyActive {
		if whereClauseStarted {
			query += " AND"
		} else {
			query += " WHERE"
		}
		query += " a1.active = true"
		whereClauseStarted = true
	}

	var params []any
	if barcode != "" {
		if whereClauseStarted {
			query += " AND"
		} else {
			query += " WHERE"
		}
		query += " a1.barcode = ?"
		params = append(params, barcode)
		whereClauseStarted = true
	}

	if !precursor {
		if whereClauseStarted {
			query += " AND"
		} else {
			query += " WHERE"
		}
		query += " a1.precursor IS NULL"
	}

	query += " ORDER BY a1.name ASC"
	rows, err := r.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	articles, err := processRows(rows)
	if err != nil {
		return nil, err
	}

	var result []domain.Article
	for _, a := range articles {
		var precursor *Article
		if a.PrecursorID.Valid {
			found, err := r.findPrecursorById(a.PrecursorID.Int64)
			if err != nil {
				return nil, err
			}
			precursor = &found
		}
		result = append(result, mapToDomain(a, precursor))
	}
	return result, nil
}

func (r Repository) FindById(ctx context.Context, aid int64) (domain.Article, error) {
	row := r.getDB(ctx).QueryRow("SELECT * FROM article WHERE id = ?", aid)

	article, err := processRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Article{}, domain.ArticleNotFoundError{Identifier: strconv.FormatInt(aid, 10)}
		}
		return domain.Article{}, err
	}
	var precursor *Article
	if article.PrecursorID.Valid {
		found, err := r.findPrecursorById(article.PrecursorID.Int64)
		if err != nil {
			return domain.Article{}, err
		}
		precursor = &found
	}
	return mapToDomain(article, precursor), nil
}

func (r Repository) FindActiveByBarcode(barcode string) (domain.Article, error) {
	row := r.db.QueryRow("SELECT * FROM article WHERE active = true and barcode = ?", barcode)
	article, err := processRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Article{}, domain.ArticleNotFoundError{Identifier: barcode}
		}
	}
	var precursor *Article
	if article.PrecursorID.Valid {
		found, err := r.findPrecursorById(article.PrecursorID.Int64)
		if err != nil {
			return domain.Article{}, err
		}
		precursor = &found
	}
	return mapToDomain(article, precursor), nil
}

func (r Repository) StoreArticle(ctx context.Context, a domain.Article) (domain.Article, error) {
	var precursorID *int64
	if a.Precursor != nil {
		precursorID = &a.Precursor.ID
	}
	a.Created = time.Now()
	res, err := r.getDB(ctx).Exec("INSERT INTO article (name, barcode, amount, active, created, usage_count, precursor_id) VALUES ($1, $2, $3, $4, $5, 0, $6)",
		a.Name, a.Barcode, a.Amount, a.IsActive, a.Created, precursorID)
	if err != nil {
		return domain.Article{}, err
	}
	a.ID, err = res.LastInsertId()
	if err != nil {
		return domain.Article{}, err
	}

	return a, nil
}

func (r Repository) UpdateArticle(ctx context.Context, a domain.Article) error {
	_, err := r.getDB(ctx).Exec("UPDATE article SET name=$1, barcode=$2, amount=$3, active=$4, usage_count=$5 WHERE ID = $6",
		a.Name, a.Barcode, a.Amount, a.IsActive, a.UsageCount, a.ID)
	return err
}

func (r Repository) DeleteById(int64) error {
	return nil
}

func (r Repository) Transactional(ctx context.Context, f func(context.Context) error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = f(database.AddToContext(ctx, tx))
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r Repository) findPrecursorById(aid int64) (Article, error) {
	row := r.db.QueryRow("SELECT * FROM article WHERE id = ?", aid)
	found, err := processRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Article{}, domain.ArticleNotFoundError{Identifier: strconv.FormatInt(aid, 10)}
		}
		return Article{}, err
	}
	return found, nil
}

func (r Repository) getDB(ctx context.Context) database.DB {
	if db, ok := database.FromContext(ctx); ok {
		return db
	}
	return r.db
}

func processRow(r *sql.Row) (Article, error) {
	var a Article
	err := r.Scan(&a.ID, &a.PrecursorID, &a.Name, &a.Barcode, &a.Amount, &a.IsActive, &a.Created, &a.UsageCount)
	if err != nil {
		return Article{}, err
	}
	return a, nil
}

func mapToDomain(a Article, precursor *Article) domain.Article {
	da := domain.Article{
		ID:         a.ID,
		Name:       a.Name,
		Amount:     a.Amount,
		IsActive:   a.IsActive,
		Created:    a.Created,
		UsageCount: a.UsageCount,
	}

	if precursor != nil {
		da.Precursor = new(domain.Article)
		*da.Precursor = mapToDomain(*precursor, nil)
	}

	if a.Barcode.Valid {
		da.Barcode = new(string)
		*da.Barcode = a.Barcode.String
	}

	return da
}

func processRows(r *sql.Rows) ([]Article, error) {
	var articles []Article
	for r.Next() {
		var a Article
		err := r.Scan(&a.ID, &a.PrecursorID, &a.Name, &a.Barcode, &a.Amount, &a.IsActive, &a.Created, &a.UsageCount)
		if err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}
