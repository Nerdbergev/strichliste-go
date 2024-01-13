package users

import (
	"errors"
	"time"

	"github.com/nerdbergev/shoppinglist-go/pkg/settings"
	"github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
)

func NewService(settings settings.Service, repo domain.UserRepository) (Service, error) {
	sp, ok := settings.Get("user.stalePeriod").(string)
	if !ok {
		return Service{}, errors.New("user.stalePeriod is not a string")
	}
	// Ignore error since the setting is validated during service start up
	dur, err := time.ParseDuration(sp)
	if err != nil {
		return Service{}, errors.New("failed to parse user.stalePeriod duration")
	}

	return Service{
		stalePeriod: dur,
		repo:        repo,
	}, nil
}

type Service struct {
	stalePeriod time.Duration
	repo        domain.UserRepository
}

func (svc Service) GetAll() ([]domain.User, error) {
	return svc.repo.GetAll()
}

type FindByStateRequest interface {
	Active() bool
}

func (svc Service) FindByState(req FindByStateRequest) ([]domain.User, error) {
	if req.Active() {
		return svc.repo.FindActive(svc.GetStaleDateTime())
	}
	return svc.repo.FindInactive(svc.GetStaleDateTime())
}

type CreateUserRequest interface {
	Name() string
	Email() string
	HasEmail() bool
}

func (svc Service) CreateUser(req CreateUserRequest) (domain.User, error) {
	u := domain.User{
		Name: req.Name(),
	}
	if req.HasEmail() {
		u.Email = new(string)
		*u.Email = req.Email()
	}
	return svc.repo.StoreUser(u)
}

func (svc Service) FindById(uid int64) (domain.User, error) {
	return svc.repo.FindById(uid)
}

func (svc Service) GetStaleDateTime() time.Time {
	return time.Now().Add(-svc.stalePeriod) // hack to substract the duration without converting to time.Time
}
