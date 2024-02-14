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

func (r Repository) GetAll(onlyActive bool) ([]domain.Article, error) {
	query := "SELECT * FROM article"
	if onlyActive {
		query += " WHERE active = true"
	}
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return processRows(rows)
}

func (r Repository) FindById(ctx context.Context, aid int64) (domain.Article, error) {
	row := r.getDB(ctx).QueryRow("SELECT * FROM article WHERE id = ?", aid)
	article, err := processRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Article{}, domain.ArticleNotFoundError{Identifier: strconv.FormatInt(aid, 10)}
		}
	}
	return article, nil
}

func (r Repository) FindByBarcode(string) (domain.Article, error) {
	return domain.Article{}, nil
}

func (r Repository) StoreArticle(a domain.Article) (domain.Article, error) {
	a.Created = time.Now()
	res, err := r.db.Exec("INSERT INTO article (name, barcode, amount, active, created, usage_count) VALUES ($1, $2, $3, $4, $5, 0)",
		a.Name, a.Barcode, a.Amount, a.IsActive, a.Created)
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

func (r Repository) getDB(ctx context.Context) database.DB {
	if db, ok := database.FromContext(ctx); ok {
		return db
	}
	return r.db
}

func processRow(r *sql.Row) (domain.Article, error) {
	var a Article
	err := r.Scan(&a.ID, &a.PrecursorID, &a.Name, &a.Barcode, &a.Amount, &a.IsActive, &a.Created, &a.UsageCount)
	if err != nil {
		return domain.Article{}, err
	}
	return mapToDomain(a), nil
}

func mapToDomain(a Article) domain.Article {
	da := domain.Article{
		ID:         a.ID,
		Name:       a.Name,
		Amount:     a.Amount,
		IsActive:   a.IsActive,
		Created:    a.Created,
		UsageCount: a.UsageCount,
	}

	if a.Barcode.Valid {
		da.Barcode = new(string)
		*da.Barcode = a.Barcode.String
	}

	return da
}

func processRows(r *sql.Rows) ([]domain.Article, error) {
	var articles []domain.Article
	for r.Next() {
		var a Article
		err := r.Scan(&a.ID, &a.PrecursorID, &a.Name, &a.Barcode, &a.Amount, &a.IsActive, &a.Created, &a.UsageCount)
		if err != nil {
			return nil, err
		}
		articles = append(articles, mapToDomain(a))
	}
	return articles, nil
}
