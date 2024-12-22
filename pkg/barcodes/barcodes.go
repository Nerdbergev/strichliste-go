package barcodes

import (
	"context"
	"errors"
	"time"

	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/barcodes/domain"
)

type Service struct {
	arepo adomain.ArticleRepository
	repo  domain.BarcodeRepository
}

func NewService(repo domain.BarcodeRepository, arepo adomain.ArticleRepository) Service {
	return Service{
		repo:  repo,
		arepo: arepo,
	}
}

func (svc Service) GetAll() ([]domain.Barcode, error) {
	return svc.repo.GetAll()
}

func (svc Service) FindByArticleID(aid int64) ([]domain.Barcode, error) {
	return svc.repo.FindByArticleID(aid)
}

func (svc Service) FindByID(bid int64) (domain.Barcode, error) {
	return svc.repo.FindByID(bid)
}

func (svc Service) AddArticleBarcode(aid int64, barcode string) (adomain.Article, error) {
	article, err := svc.arepo.FindById(context.Background(), aid)
	if err != nil {
		return adomain.Article{}, err
	}

	_, err = svc.repo.FindByBarcode(barcode)
	if err == nil {
		return adomain.Article{}, errors.New("barcode already exists")
	} else {
		var bnf domain.BarcodeNotFoundError
		if !errors.As(err, &bnf) {
			return adomain.Article{}, err
		}
	}

	bc := domain.Barcode{
		ArticleID: article.ID,
		Barcode:   barcode,
		Created:   time.Now(),
	}

	stored, err := svc.repo.StoreBarcode(bc)
	if err != nil {
		return adomain.Article{}, err
	}

	article.Barcodes = append(article.Barcodes, adomain.Barcode{
		ID:      stored.ID,
		Barcode: stored.Barcode,
		Created: stored.Created,
	})
	return article, nil
}

func (svc Service) DeleteArticleBarcode(aid, bid int64) (adomain.Article, error) {
	article, err := svc.arepo.FindById(context.Background(), aid)
	if err != nil {
		return adomain.Article{}, err
	}

	_, err = svc.repo.FindByID(bid)
	if err != nil {
		return adomain.Article{}, err
	}

	err = svc.repo.DeleteByID(bid)
	if err != nil {
		return adomain.Article{}, err
	}

	for i := range article.Barcodes {
		if article.Barcodes[i].ID == bid {
			article.Barcodes = append(article.Barcodes[:i], article.Barcodes[i+1:]...)
			break
		}
	}

	return article, nil
}
