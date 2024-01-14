package articles

import "github.com/nerdbergev/shoppinglist-go/pkg/articles/domain"

func NewService(repo domain.ArticleRepository) Service {
	return Service{repo: repo}
}

type Service struct {
	repo domain.ArticleRepository
}

func (svc Service) GetAll() ([]domain.Article, error) {
	return svc.repo.GetAll()
}
