package transactions

import (
	"context"
	"errors"
	"time"

	adomain "github.com/nerdbergev/shoppinglist-go/pkg/articles/domain"
	"github.com/nerdbergev/shoppinglist-go/pkg/settings"
	"github.com/nerdbergev/shoppinglist-go/pkg/transactions/domain"
	udomain "github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
)

var (
	ErrTransactionInvalid = errors.New("Amout can't be positive when sending money or buying an article")
	ErrUserNotFound       = errors.New("User not found")
)

func NewService(repo domain.TransactionRepository, urepo udomain.UserRepository, arepo adomain.ArticleRepository, settings settings.Service) Service {
	return Service{
		repo:     repo,
		urepo:    urepo,
		arepo:    arepo,
		settings: settings,
	}
}

type Service struct {
	repo     domain.TransactionRepository
	urepo    udomain.UserRepository
	arepo    adomain.ArticleRepository
	settings settings.Service
}

func (svc Service) GetFromUser(uid int64) ([]domain.Transaction, error) {
	return svc.repo.FindByUserId(uid)
}

func (svc Service) ProcessTransaction(uid, amount int64, comment *string, quantity, articleID, recipientID *int64) (domain.Transaction, error) {
	ctx := context.Background()
	var processed domain.Transaction
	err := svc.repo.Transaction(ctx, func(ctx context.Context) error {
		if (recipientID != nil || articleID != nil) && amount > 0 {
			return ErrTransactionInvalid
		}

		user, err := svc.urepo.FindById(ctx, uid)
		if err != nil {
			return err
		}

		t := domain.Transaction{
			User:    user,
			Comment: comment,
			Created: time.Now(),
		}

		if articleID != nil {
			article, err := svc.arepo.FindById(ctx, *articleID)
			if err != nil {
				return err
			}
			if !article.IsActive {
				return errors.New("article inactive")
			}
			t.Article = &article
			if quantity == nil {
				quantity = new(int64)
				*quantity = 1
			}
			t.Quantity = quantity
			amount = article.Amount * *t.Quantity * -1
			article.IncrementUsageCount()
			if err := svc.arepo.UpdateArticle(ctx, article); err != nil {
				return err
			}
		}

		t.Amount = amount
		if err := svc.checkTransactionBoundary(amount); err != nil {
			return err
		}

		user.AddBalance(amount)
		if err := svc.checkAccountBalanceBoundary(user); err != nil {
			return err
		}

		processed, err = svc.repo.StoreTransaction(ctx, t)
		if err != nil {
			return err
		}

		err = svc.urepo.UpdateUser(ctx, user)
		if err != nil {
			return err
		}
		processed.User = user
		return nil
	})
	if err != nil {
		return domain.Transaction{}, err
	}
	return processed, nil
}

func (svc Service) checkAccountBalanceBoundary(user udomain.User) error {
	upper, ok := svc.settings.GetInt("account.boundary.upper")
	if ok && int64(upper) < user.Balance {
		return errors.New("Account Balance error")
	}

	lower, ok := svc.settings.GetInt("account.boundary.lower")
	if ok && user.Balance < int64(lower) {
		return errors.New("Account Balance error")
	}

	return nil
}

func (svc Service) checkTransactionBoundary(amount int64) error {
	upper, ok := svc.settings.GetInt("payment.boundary.upper")
	if ok && int64(upper) < amount {
		return errors.New("Transaction Boundary error")
	}

	lower, ok := svc.settings.GetInt("payment.boundary.lower")
	if ok && amount < int64(lower) {
		return errors.New("Transaction Boundary error")
	}
	return nil
}
