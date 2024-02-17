package articles

import (
	"context"
	"errors"

	"github.com/nerdbergev/strichliste-go/pkg/articles/domain"
)

func NewService(repo domain.ArticleRepository) Service {
	return Service{repo: repo}
}

type Service struct {
	repo domain.ArticleRepository
}

type Filter interface {
	Value() any
}

func (svc Service) GetAll(onlyActive, precursor bool, barcode string, ancestor *bool) ([]domain.Article, error) {
	return svc.repo.GetAll(onlyActive, precursor, barcode, ancestor)
}

type ArticleRequest interface {
	Name() string
	HasBarcode() bool
	Barcode() string
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
	if req.HasBarcode() {
		if existing, err := svc.repo.FindActiveByBarcode(req.Barcode()); err == nil {
			return domain.Article{}, domain.ArticleBarcodeAlreadyExistsError{
				Id:      existing.ID,
				Barcode: *existing.Barcode,
			}
		}
		a.Barcode = new(string)
		*a.Barcode = req.Barcode()
	}

	return svc.repo.StoreArticle(context.Background(), a)
}

func (svc Service) UpdateArticle(aid int64, req ArticleRequest) (domain.Article, error) {
	exists, err := svc.repo.FindById(context.Background(), aid)
	if err != nil {
		return domain.Article{}, err
	}
	if !exists.IsActive {
		return domain.Article{}, domain.ArticleInactiveError{Id: aid, Name: exists.Name}
	}

	newArticle := domain.Article{
		Name:     req.Name(),
		IsActive: req.IsActive(),
		Amount:   req.Amount(),
	}

	if req.HasBarcode() {
		byBarcode, err := svc.repo.FindActiveByBarcode(req.Barcode())
		if err == nil {
			if byBarcode.ID != exists.ID {
				return domain.Article{}, domain.ArticleBarcodeAlreadyExistsError{
					Id:      byBarcode.ID,
					Barcode: *byBarcode.Barcode,
				}
			}
		} else if !errors.As(err, &domain.ArticleNotFoundError{}) {
			return domain.Article{}, err
		}
		newArticle.Barcode = new(string)
		*newArticle.Barcode = req.Barcode()
	}

	newArticle.UsageCount = exists.UsageCount
	exists.IsActive = false

	var updated domain.Article
	if err := svc.repo.Transactional(context.Background(), func(ctx context.Context) error {
		if err := svc.repo.UpdateArticle(ctx, exists); err != nil {
			return err
		}
		newArticle.Precursor = &exists
		updated, err = svc.repo.StoreArticle(ctx, newArticle)
		if err != nil {
			return err
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
