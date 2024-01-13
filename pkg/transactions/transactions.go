package transactions

import (
	"context"
	"errors"

	"github.com/nerdbergev/shoppinglist-go/pkg/transactions/domain"
	udomain "github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
)

var (
	ErrTransactionInvalid = errors.New("Amout can't be positive when sending money or buying an article")
	ErrUserNotFound       = errors.New("User not found")
)

func NewService(repo domain.TransactionRepository, urepo udomain.UserRepository) Service {
	return Service{
		repo:  repo,
		urepo: urepo,
	}
}

type Service struct {
	repo  domain.TransactionRepository
	urepo udomain.UserRepository
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

		// if articleID != nil {

		// }

		return nil
	})
	if err != nil {
		return domain.Transaction{}, err
	}
	return t, nil
}
