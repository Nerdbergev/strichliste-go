package users

import (
	"time"

	"github.com/nerdbergev/shoppinglist-go/pkg/settings"
)

type Service struct {
	settings settings.Service
}

func (svc Service) GetStaleDateTime() time.Time {
	svc.settings.Get("user.stalePeriod")
	return time.Now()
}
