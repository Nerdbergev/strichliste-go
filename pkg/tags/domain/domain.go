package domain

import (
	"context"
	"time"
)

type Tag struct {
	ID         int64
	Tag        string
	Created    time.Time
	UsageCount int64
}

type TagRepository interface {
	GetAll() ([]Tag, error)
	GetArticleTags(int64) ([]Tag, error)
	FindByID(int64) (Tag, error)
	FindByTag(string) (Tag, error)
	CheckArticleHasTag(int64, int64) bool
	Transactional(context.Context, func(context.Context) error) error
	CreateTag(context.Context, string, time.Time) (Tag, error)
	AddArticleTag(context.Context, int64, int64, time.Time) error
}
