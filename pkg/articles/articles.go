package articles

import (
	"context"

	"github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	tdomain "github.com/nerdbergev/strichliste-go/pkg/transactions/domain"
)

func NewService(repo domain.ArticleRepository,
	trepo tdomain.TransactionRepository,
	brepo domain.BarcodeRepository, tagrepo domain.TagRepository) Service {
	return Service{repo: repo, trepo: trepo, brepo: brepo, tagrepo: tagrepo}
}

type Service struct {
	repo    domain.ArticleRepository
	trepo   tdomain.TransactionRepository
	tagrepo domain.TagRepository
	brepo   domain.BarcodeRepository
}

func (svc Service) GetAll(onlyActive, precursor bool, barcode string, ancestor *bool) ([]domain.Article, error) {
	articles, err := svc.repo.GetAll(onlyActive, precursor, barcode, ancestor)
	if err != nil {
		return nil, err
	}

	for i := range articles {
		aid := articles[i].ID
		barcodes, err := svc.brepo.FindByArticleID(aid)
		if err != nil {
			return nil, err
		}
		articles[i].Barcodes = barcodes

		tags, err := svc.tagrepo.GetArticleTags(aid)
		if err != nil {
			return nil, err
		}
		articles[i].Tags = tags
	}
	return articles, nil
}

func (svc Service) CountActive() (int, error) {
	return svc.repo.CountActive()
}

func (svc Service) FindById(aid domain.ArticleID) (domain.Article, error) {
	article, err := svc.repo.FindById(context.Background(), aid)
	if err != nil {
		return domain.Article{}, err
	}

	barcodes, err := svc.brepo.FindByArticleID(article.ID)
	if err != nil {
		return domain.Article{}, err
	}
	article.Barcodes = barcodes

	tags, err := svc.tagrepo.GetArticleTags(article.ID)
	if err != nil {
		return domain.Article{}, err
	}
	article.Tags = tags

	return article, err
}

type ArticleRequest interface {
	Name() string
	IsActive() bool
	Amount() int64
	HasPrecursor() bool
	Precursor() ArticleRequest
}

func (svc Service) CreateArticle(req ArticleRequest) (domain.Article, error) {
	a := domain.Article{
		Name:     req.Name(),
		IsActive: req.IsActive(),
		Amount:   req.Amount(),
	}

	return svc.repo.Store(context.Background(), a)
}

func (svc Service) UpdateArticle(aid domain.ArticleID, req ArticleRequest) (domain.Article, error) {
	existing, err := svc.repo.FindById(context.Background(), aid)
	if err != nil {
		return domain.Article{}, err
	}

	if !existing.IsActive {
		return domain.Article{}, domain.ArticleInactiveError{Id: existing.ID, Name: existing.Name}
	}

	// We could probably just use the usage count from the existing article but the original php
	// backend implemented it like this and since one goal of the go rewrite is to be bug and
	// feature compatible, we'll just do the same.
	referenceCount, err := svc.trepo.GetArticleReferenceCount(int64(aid))
	if err != nil {
		return domain.Article{}, err
	}

	// Article was not used before, just update the fields
	if referenceCount == 0 {

		existing.Name = req.Name()
		existing.Amount = req.Amount()

		if existing.IsActivatable() && req.IsActive() {
			existing.IsActive = true
		}

		err = svc.repo.Update(context.Background(), existing)
		return existing, err
	}

	newArticle := domain.Article{
		Name:     req.Name(),
		IsActive: req.IsActive(),
		Amount:   req.Amount(),
	}

	newArticle.UsageCount = existing.UsageCount
	existing.IsActive = false

	var updated domain.Article
	if err := svc.repo.Transactional(context.Background(), func(ctx context.Context) error {

		if err := svc.repo.Update(ctx, existing); err != nil {
			return err
		}
		newArticle.Precursor = &existing
		updated, err = svc.repo.Store(ctx, newArticle)
		if err != nil {
			return err
		}
		for _, bc := range existing.Barcodes {
			if err := svc.brepo.ReassignBarcode(ctx, bc.ID, updated.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return domain.Article{}, err
	}

	return updated, nil
}

func (svc Service) DeactivateArticle(aid domain.ArticleID) (domain.Article, error) {
	article, err := svc.repo.FindById(context.Background(), aid)
	if err != nil {
		return domain.Article{}, err
	}

	article.IsActive = false

	if err = svc.repo.Update(context.Background(), article); err != nil {
		return domain.Article{}, err
	}
	return article, nil
}
