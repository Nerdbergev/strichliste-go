package tags

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"time"

	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/tags/domain"
)

type Service struct {
	repo  domain.TagRepository
	arepo adomain.ArticleRepository
}

func NewService(repo domain.TagRepository, arepo adomain.ArticleRepository) Service {
	return Service{
		repo:  repo,
		arepo: arepo,
	}
}

func (svc Service) GetAll() ([]domain.Tag, error) {
	tags, err := svc.repo.GetAll()
	if err != nil {
		return nil, err
	}
	slices.SortFunc(tags, func(a, b domain.Tag) int {
		return cmp.Or(
			cmp.Compare(a.UsageCount, b.UsageCount),
			a.Created.Compare(b.Created),
		)
	})
	return tags, nil
}

func (svc Service) GetArticleTags(aid int64) ([]domain.Tag, error) {
	return svc.repo.GetArticleTags(aid)
}

func (svc Service) FindByID(tid int64) (domain.Tag, error) {
	return svc.repo.FindByID(tid)
}

func (svc Service) AddArticleTag(aid int64, tag string) (adomain.Article, error) {
	article, err := svc.arepo.FindById(context.Background(), aid)
	if err != nil {
		return adomain.Article{}, err
	}

	existing, err := svc.repo.FindByTag(tag)
	if err == nil {
		if svc.repo.CheckArticleHasTag(aid, existing.ID) {
			// TODO: return domain error
			return adomain.Article{}, errors.New("article has tag")
		}

		err := svc.repo.AddArticleTag(context.Background(), article.ID, existing.ID, time.Now())
		if err != nil {
			return adomain.Article{}, err
		}
		// TODO article add tag
		return article, nil
	}

	err = svc.repo.Transactional(context.Background(), func(ctx context.Context) error {
		created := time.Now()
		newTag, err := svc.repo.CreateTag(context.Background(), tag, created)
		if err != nil {
			return err
		}

		err = svc.repo.AddArticleTag(context.Background(), article.ID, newTag.ID, created)
		if err != nil {
			return err
		}

		// TODO: article add newTag
		return nil
	})
	if err != nil {
		return adomain.Article{}, err
	}

	return article, nil
}
