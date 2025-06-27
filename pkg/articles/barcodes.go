package articles

import (
	"errors"
	"slices"
	"time"

	"github.com/nerdbergev/strichliste-go/pkg/articles/domain"
)

func (svc Service) GetAllBarcodes() ([]domain.Barcode, error) {
	return svc.brepo.GetAll()
}

func (svc Service) FindBarcodeByArticleID(aid domain.ArticleID) ([]domain.Barcode, error) {
	return svc.brepo.FindByArticleID(aid)
}

func (svc Service) FindBarcodeByID(bid domain.BarcodeID) (domain.Barcode, error) {
	return svc.brepo.FindByID(bid)
}

func (svc Service) AddArticleBarcode(aid domain.ArticleID, barcode string) (domain.Article, error) {
	article, err := svc.FindById(aid)
	// article, err := svc.repo.FindById(context.Background(), aid)
	if err != nil {
		return domain.Article{}, err
	}

	_, err = svc.brepo.FindByBarcode(barcode)
	if err == nil {
		return domain.Article{}, errors.New("barcode already exists")
	} else {
		var bnf domain.BarcodeNotFoundError
		if !errors.As(err, &bnf) {
			return domain.Article{}, err
		}
	}

	bc := domain.Barcode{
		ArticleID: article.ID,
		Barcode:   barcode,
		Created:   time.Now(),
	}

	stored, err := svc.brepo.Store(bc)
	if err != nil {
		return domain.Article{}, err
	}

	article.Barcodes = append(article.Barcodes, domain.Barcode{
		ID:      stored.ID,
		Barcode: stored.Barcode,
		Created: stored.Created,
	})
	return article, nil
}

func (svc Service) DeleteArticleBarcode(aid domain.ArticleID, bid domain.BarcodeID) (domain.Article, error) {
	article, err := svc.FindById(aid)
	if err != nil {
		return domain.Article{}, err
	}

	_, err = svc.brepo.FindByID(bid)
	if err != nil {
		return domain.Article{}, err
	}

	err = svc.brepo.DeleteByID(bid)
	if err != nil {
		return domain.Article{}, err
	}

	for i := range article.Barcodes {
		if article.Barcodes[i].ID == bid {
			article.Barcodes = slices.Delete(article.Barcodes, i, i+1)
			break
		}
	}

	return article, nil
}
