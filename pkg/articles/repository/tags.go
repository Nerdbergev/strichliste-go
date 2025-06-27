package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/database"
)

type Tag struct {
	ID         sql.NullInt64
	Tag        sql.NullString
	Created    sql.NullTime
	UsageCount int64
}

func NewTagRepository(db *sql.DB) TagRepository {
	return TagRepository{db: db}
}

type TagRepository struct {
	db *sql.DB
}

func (r TagRepository) GetAll() ([]domain.Tag, error) {
	rows, err := r.db.Query(`SELECT t.id, t.tag, t.created, count(at.id) FROM tag t 
    LEFT JOIN article_tag at ON (at.tag_id = t.id) GROUP BY t.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return processTagRows(rows)
}

func (r TagRepository) GetArticleTags(articleID domain.ArticleID) ([]domain.Tag, error) {
	rows, err := r.db.Query(`SELECT t.id, t.tag, t.created, count(at.id) FROM tag t 
    LEFT JOIN article_tag at ON (at.tag_id = t.id) WHERE at.article_id = ? GROUP BY t.id`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return processTagRows(rows)
}

func (r TagRepository) FindByID(tid domain.TagID) (domain.Tag, error) {
	row := r.db.QueryRow(`SELECT t.id, t.tag, t.created, count(at.id) FROM tag t 
    LEFT JOIN article_tag at ON (at.tag_id = t.id) WHERE t.id = ?`, tid)

	var t Tag
	if err := row.Scan(&t.ID, &t.Tag, &t.Created, &t.UsageCount); err != nil {
		return domain.Tag{}, err
	}

	var dt domain.Tag
	if t.ID.Valid {
		dt = domain.Tag{
			ID:         domain.TagID(t.ID.Int64),
			Tag:        t.Tag.String,
			Created:    t.Created.Time,
			UsageCount: t.UsageCount,
		}
	}
	return dt, nil
}

func (r TagRepository) Transactional(ctx context.Context, fn func(context.Context) error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = fn(database.AddToContext(ctx, tx))
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r TagRepository) CheckArticleHasTag(aid domain.ArticleID, tid domain.TagID) bool {
	row := r.db.QueryRow("SELECT id FROM article_tag WHERE article_id = ? and tag_id = ?", aid, tid)

	var id sql.NullInt64
	if err := row.Scan(&id); err != nil {
		return false
	}
	return id.Valid
}

func (r TagRepository) FindByTag(tag string) (domain.Tag, error) {
	row := r.db.QueryRow(`SELECT t.id, t.tag, t.created, count(at.id) FROM tag t 
    LEFT JOIN article_tag at ON (at.tag_id = t.id) WHERE t.tag = ?`, tag)

	var t Tag
	if err := row.Scan(&t.ID, &t.Tag, &t.Created, &t.UsageCount); err != nil {
		return domain.Tag{}, err
	}

	var dt domain.Tag
	if t.ID.Valid {
		dt = domain.Tag{
			ID:         domain.TagID(t.ID.Int64),
			Tag:        t.Tag.String,
			Created:    t.Created.Time,
			UsageCount: t.UsageCount,
		}
	} else {
		return domain.Tag{}, errors.New("not found")
	}
	return dt, nil
}

func (r TagRepository) CreateTag(ctx context.Context, tag string, created time.Time) (domain.Tag, error) {
	result, err := r.getDB(ctx).Exec("INSERT INTO tag (tag, created) VALUES (?, ?)", tag, created)
	if err != nil {
		return domain.Tag{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Tag{}, err
	}

	return domain.Tag{ID: domain.TagID(id), Tag: tag, Created: created}, nil
}

func (r TagRepository) AddArticleTag(ctx context.Context, aid domain.ArticleID, tid domain.TagID, created time.Time) error {
	_, err := r.getDB(ctx).Exec(`INSERT INTO article_tag (article_id, tag_id, created) 
        VALUES (?, ?, ?)`, aid, tid, created)
	return err
}

func (r TagRepository) getDB(ctx context.Context) database.DB {
	if db, ok := database.FromContext(ctx); ok {
		return db
	}
	return r.db
}

func processTagRows(r *sql.Rows) ([]domain.Tag, error) {
	var tags []domain.Tag
	for r.Next() {
		var t Tag
		err := r.Scan(&t.ID, &t.Tag, &t.Created, &t.UsageCount)
		if err != nil {
			return nil, err
		}
		if t.ID.Valid {
			tags = append(tags, domain.Tag{
				ID:         domain.TagID(t.ID.Int64),
				Tag:        t.Tag.String,
				Created:    t.Created.Time,
				UsageCount: t.UsageCount,
			})
		}
	}
	return tags, nil
}
