package articles

import "github.com/nerdbergev/shoppinglist-go/pkg/articles/domain"

func NewService(repo domain.ArticleRepository) Service {
	return Service{repo: repo}
}

type Service struct {
	repo domain.ArticleRepository
}

func (svc Service) GetAll(onlyActive bool) ([]domain.Article, error) {
	return svc.repo.GetAll(onlyActive)
}

type CreateArticleRequest interface {
	Name() string
	HasBarcode() bool
	Barcode() string
	IsActive() bool
	Amount() int64
}

func (svc Service) CreateArticle(req CreateArticleRequest) (domain.Article, error) {
	a := domain.Article{
		Name:     req.Name(),
		IsActive: req.IsActive(),
		Amount:   req.Amount(),
	}
	if req.HasBarcode() {
		a.Barcode = new(string)
		*a.Barcode = req.Barcode()
	}
	return svc.repo.StoreArticle(a)
}
