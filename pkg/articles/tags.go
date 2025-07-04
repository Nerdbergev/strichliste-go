package articles

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"time"

	domain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
)

func (svc Service) GetAllTags() ([]domain.Tag, error) {
	tags, err := svc.tagrepo.GetAll()
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

func (svc Service) GetArticleTags(aid domain.ArticleID) ([]domain.Tag, error) {
	return svc.tagrepo.GetArticleTags(aid)
}

func (svc Service) FindTagByID(tid domain.TagID) (domain.Tag, error) {
	return svc.tagrepo.FindByID(tid)
}

func (svc Service) AddArticleTag(aid domain.ArticleID, tag string) (domain.Article, error) {
	article, err := svc.FindById(aid)
	if err != nil {
		return domain.Article{}, err
	}

	existing, err := svc.tagrepo.FindByTag(tag)
	if err == nil {
		if svc.tagrepo.CheckArticleHasTag(aid, existing.ID) {
			// TODO: return domain error
			return domain.Article{}, errors.New("article has tag")
		}

		err := svc.tagrepo.AddArticleTag(context.Background(), article.ID, existing.ID, time.Now())
		if err != nil {
			return domain.Article{}, err
		}
		// TODO article add tag
		article.Tags = append(article.Tags, domain.Tag{
			ID:  existing.ID,
			Tag: tag,
		})
		return article, nil
	}

	err = svc.repo.Transactional(context.Background(), func(ctx context.Context) error {
		created := time.Now()
		newTag, err := svc.tagrepo.CreateTag(context.Background(), tag, created)
		if err != nil {
			return err
		}

		err = svc.tagrepo.AddArticleTag(context.Background(), article.ID, newTag.ID, created)
		if err != nil {
			return err
		}

		// TODO: article add newTag
		article.Tags = append(article.Tags, newTag)
		return nil
	})
	if err != nil {
		return domain.Article{}, err
	}

	return article, nil
}

func (svc Service) DeleteArticleTag(aid domain.ArticleID, tid domain.TagID) (domain.Article, error) {
	tag, err := svc.tagrepo.FindByID(tid)
	if err != nil {
		return domain.Article{}, err
	}

	err = svc.tagrepo.Transactional(context.Background(), func(ctx context.Context) error {
		err := svc.tagrepo.DeleteArticleTag(ctx, aid, tid)
		if err != nil {
			return err
		}

		if tag.UsageCount == 1 {
			err := svc.tagrepo.DeleteTag(ctx, tid)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return domain.Article{}, err
	}

	return svc.FindById(aid)
}
