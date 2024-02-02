package transactions

import (
	"context"
	"errors"
	"time"

	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/settings"
	"github.com/nerdbergev/strichliste-go/pkg/transactions/domain"
	udomain "github.com/nerdbergev/strichliste-go/pkg/users/domain"
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
	transactions, err := svc.repo.FindByUserId(uid)
	if err != nil {
		return nil, err
	}
	for i := range transactions {
		transactions[i].IsDeletable = svc.isDeletable(transactions[i])
	}
	return transactions, nil
}

func (svc Service) ProcessTransaction(uid, amount int64, comment *string, quantity, articleID, recipientID *int64) (domain.Transaction, error) {
	var processed domain.Transaction
	err := svc.repo.Transactional(context.Background(), func(ctx context.Context) error {
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

		var article adomain.Article
		if articleID != nil {
			article, err = svc.arepo.FindById(ctx, *articleID)
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

		var recipientTransaction domain.Transaction
		if recipientID != nil {
			recipient, err := svc.urepo.FindById(ctx, *recipientID)
			if err != nil {
				return err
			}

			recipientTransaction = domain.Transaction{
				Amount:  amount * -1,
				Comment: comment,
				User:    recipient,
				Created: time.Now(),
			}

			if articleID != nil {
				recipientTransaction.Article = &article
			}

			recipient.AddBalance(amount * -1)
			if err := svc.checkAccountBalanceBoundary(recipient); err != nil {
				return err
			}

			recipientTransaction, err = svc.repo.StoreTransaction(ctx, recipientTransaction)
			if err != nil {
				return err
			}
			t.RecipientTransaction = &recipientTransaction

			if err := svc.urepo.UpdateUser(ctx, recipient); err != nil {
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
		if recipientID != nil {
			recipientTransaction.SenderTransaction = &processed
			if err := svc.repo.UpdateSenderTransaction(ctx, recipientTransaction); err != nil {
				return err
			}
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
	processed.IsDeletable = svc.isDeletable(processed)
	return processed, nil
}

func (svc Service) RevertTransaction(tid int64) (domain.Transaction, error) {
	var reverted domain.Transaction
	if err := svc.repo.Transactional(context.Background(), func(ctx context.Context) error {
		toRevert, err := svc.repo.FindById(ctx, tid)
		if err != nil {
			return err
		}

		if toRevert.Article != nil {
			toRevert.Article.DecrementUsageCount()
			if err := svc.arepo.UpdateArticle(ctx, *toRevert.Article); err != nil {
				return err
			}
		}

		if toRevert.RecipientTransaction != nil {
			if err := svc.undoTransaction(ctx, *toRevert.RecipientTransaction); err != nil {
				return err
			}
		}

		if toRevert.SenderTransaction != nil {
			if err := svc.undoTransaction(ctx, *toRevert.SenderTransaction); err != nil {
				return err
			}
		}
		reverted = toRevert
		return svc.undoTransaction(ctx, toRevert)
	}); err != nil {
		return domain.Transaction{}, err
	}
	reverted.IsDeleted = true
	return reverted, nil
}

func (svc Service) undoTransaction(ctx context.Context, t domain.Transaction) error {
	user := t.User
	if err := svc.checkTransactionBoundary(t.Amount); err != nil {
		return err
	}

	user.AddBalance(t.Amount * -1)
	if err := svc.checkAccountBalanceBoundary(user); err != nil {
		return err
	}

	undoEnabled, ok := svc.settings.GetBool("payment.undo.delete")
	if ok && undoEnabled {
		if err := svc.repo.DeleteById(ctx, t.ID); err != nil {
			return err
		}
	} else {
		t.IsDeleted = true
		if err := svc.repo.MarkDeleted(ctx, t.ID); err != nil {
			return err
		}
	}

	return svc.urepo.UpdateUser(ctx, user)
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

func (svc Service) isDeletable(t domain.Transaction) bool {
	if t.IsDeleted {
		return false
	}

	undoEnabled, ok := svc.settings.GetBool("payment.undo.enabled")
	if !ok || !undoEnabled {
		return false
	}

	deletionTimeout, ok := svc.settings.GetString("payment.undo.timeout")
	if !ok {
		return false
	}

	dur, err := time.ParseDuration(deletionTimeout)
	if err != nil {
		return false
	}

	return time.Now().Before(t.Created.Add(dur))
}
