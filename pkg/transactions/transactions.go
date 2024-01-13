package transactions

import (
	"github.com/nerdbergev/shoppinglist-go/pkg/transactions/domain"
)

func NewService(repo domain.TransactionRepository) Service {
	return Service{
		repo: repo,
	}
}

type Service struct {
	repo domain.TransactionRepository
}

func (svc Service) GetFromUser(uid int64) ([]domain.Transaction, error) {
	return svc.repo.FindByUserId(uid)
}
