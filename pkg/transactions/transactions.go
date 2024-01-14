package transactions

import (
	"context"
	"errors"

	adomain "github.com/nerdbergev/shoppinglist-go/pkg/articles/domain"
	"github.com/nerdbergev/shoppinglist-go/pkg/transactions/domain"
	udomain "github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
)

var (
	ErrTransactionInvalid = errors.New("Amout can't be positive when sending money or buying an article")
	ErrUserNotFound       = errors.New("User not found")
)

func NewService(repo domain.TransactionRepository, urepo udomain.UserRepository, arepo adomain.ArticleRepository) Service {
	return Service{
		repo:  repo,
		urepo: urepo,
		arepo: arepo,
	}
}

type Service struct {
	repo  domain.TransactionRepository
	urepo udomain.UserRepository
	arepo adomain.ArticleRepository
}

func (svc Service) GetFromUser(uid int64) ([]domain.Transaction, error) {
	return svc.repo.FindByUserId(uid)
}

func (svc Service) ProcessTransaction(uid, amount int64, comment *string, quantity, articleID, recipientID *int64) (domain.Transaction, error) {
	ctx := context.Background()
	var t domain.Transaction
	err := svc.repo.Transaction(ctx, func(ctx context.Context) error {
		if (recipientID != nil || articleID != nil) && amount > 0 {
			return ErrTransactionInvalid
		}

		user, err := svc.urepo.FindById(ctx, uid)
		if err != nil {
			return err
		}

		t = domain.Transaction{
			User:    user,
			Comment: comment,
		}

		if articleID != nil {
			article, err := svc.arepo.FindById(ctx, *articleID)
			if err != nil {
				return err
			}
			t.Article = &article

		}

		return nil
	})
	if err != nil {
		return domain.Transaction{}, err
	}
	return t, nil
}
