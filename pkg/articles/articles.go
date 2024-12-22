package articles

import (
	"context"

	"github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	bdomain "github.com/nerdbergev/strichliste-go/pkg/barcodes/domain"
	tdomain "github.com/nerdbergev/strichliste-go/pkg/transactions/domain"
)

func NewService(repo domain.ArticleRepository,
	trepo tdomain.TransactionRepository,
	brepo bdomain.BarcodeRepository) Service {
	return Service{repo: repo, trepo: trepo, brepo: brepo}
}

type Service struct {
	repo  domain.ArticleRepository
	trepo tdomain.TransactionRepository
	brepo bdomain.BarcodeRepository
}

type Filter interface {
	Value() any
}

func (svc Service) GetAll(onlyActive, precursor bool, barcode string, ancestor *bool) ([]domain.Article, error) {
	return svc.repo.GetAll(onlyActive, precursor, barcode, ancestor)
}

func (svc Service) CountActive() (int, error) {
	return svc.repo.CountActive()
}

func (svc Service) FindById(aid int64) (domain.Article, error) {
	return svc.repo.FindById(context.Background(), aid)
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

	return svc.repo.StoreArticle(context.Background(), a)
}

func (svc Service) UpdateArticle(aid int64, req ArticleRequest) (domain.Article, error) {
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
	referenceCount, err := svc.trepo.GetArticleReferenceCount(aid)
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

		err = svc.repo.UpdateArticle(context.Background(), existing)
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

		if err := svc.repo.UpdateArticle(ctx, existing); err != nil {
			return err
		}
		newArticle.Precursor = &existing
		updated, err = svc.repo.StoreArticle(ctx, newArticle)
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

func (svc Service) DeactivateArticle(aid int64) (domain.Article, error) {
	article, err := svc.repo.FindById(context.Background(), aid)
	if err != nil {
		return domain.Article{}, err
	}

	article.IsActive = false

	if err = svc.repo.UpdateArticle(context.Background(), article); err != nil {
		return domain.Article{}, err
	}
	return article, nil
}
