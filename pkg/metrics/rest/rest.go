package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/render"
	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/metrics"
	"github.com/nerdbergev/strichliste-go/pkg/metrics/domain"
)

type Handler struct {
	svc metrics.Service
}

func NewHandler(svc metrics.Service) Handler {
	return Handler{svc: svc}
}

func (h Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	daysParam := strings.TrimSpace(r.URL.Query().Get("days"))
	days := 30
	if daysParam != "" {
		var err error
		days, err = strconv.Atoi(daysParam)
		if err != nil {
			_ = render.Render(w, r, ErrRender(err))
			return
		}
	}

	balance, err := h.svc.GetBalance()
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	transactionCount, err := h.svc.GetTransactionCount()
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	userCount, err := h.svc.GetUserCount()
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	articles, err := h.svc.GetArticles()
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	perDay, err := h.svc.GetTransactionsPerDay(days)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	mr := MetricsResponse{
		Balance:          balance,
		TransactionCount: transactionCount,
		UserCount:        userCount,
		Articles:         MapArticles(articles),
		Days:             MapDays(perDay),
	}

	render.JSON(w, r, mr)
}

func (h Handler) GetUserMetrics(w http.ResponseWriter, r *http.Request) {

}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

type MetricsResponse struct {
	Balance          int64     `json:"balance"`
	TransactionCount int64     `json:"transactionCount"`
	UserCount        int       `json:"userCount"`
	Articles         []Article `json:"articles"`
	Days             []Day     `json:"days"`
}

func (mr MetricsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type Barcode struct {
	ID      int64     `json:"id"`
	Barcode string    `json:"barcode"`
	Created time.Time `json:"created"`
}

type Article struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Barcodes   []Barcode `json:"barcodes"`
	Amount     int64     `json:"amount"`
	IsActive   bool      `json:"isActive"`
	UsageCount int64     `json:"usageCount"`
	Precursor  *Article  `json:"precursor"`
	Created    time.Time `json:"created"`
}

type TurnoverOrZero struct {
	t *Turnover
}

func (tz *TurnoverOrZero) MarshalJSON() ([]byte, error) {
	if tz.t == nil {
		return []byte("0"), nil
	}
	return json.Marshal(tz.t)
}

type Turnover struct {
	Amount            int `json:"amount"`
	TransactionsCount int `json:"transactions"`
}

type Day struct {
	Date              string         `json:"date"`
	TransactionCount  int            `json:"transactions"`
	DistinctUserCount int            `json:"distinctUsers"`
	Balance           int64          `json:"balance"`
	Charged           TurnoverOrZero `json:"charged,omitempty"`
	Spent             TurnoverOrZero `json:"spent,omitempty"`
}

func MapArticles(articles []adomain.Article) []Article {
	var mapped []Article
	for _, a := range articles {
		mapped = append(mapped, MapArticle(a))
	}
	return mapped
}

func MapArticle(article adomain.Article) Article {
	resp := Article{
		ID:         article.ID,
		Name:       article.Name,
		Amount:     article.Amount,
		IsActive:   article.IsActive,
		UsageCount: article.UsageCount,
		Created:    article.Created,
	}

	for _, bc := range article.Barcodes {
		resp.Barcodes = append(resp.Barcodes, Barcode{
			ID:      bc.ID,
			Barcode: bc.Barcode,
			Created: bc.Created,
		})
	}

	if article.Precursor != nil {
		resp.Precursor = new(Article)
		*resp.Precursor = MapArticle(*article.Precursor)
	}
	return resp
}

func MapDays(days []domain.Day) []Day {
	var mapped []Day
	for _, d := range days {
		mapped = append(mapped, MapDay(d))
	}
	return mapped
}

func MapDay(day domain.Day) Day {
	mapped := Day{
		Date:              day.Date,
		TransactionCount:  day.TransactionCount,
		DistinctUserCount: day.DistinctUserCount,
		Balance:           day.Balance,
	}

	if day.Charged != nil {
		mapped.Charged = TurnoverOrZero{
			t: &Turnover{
				Amount:            day.Charged.Amount,
				TransactionsCount: day.Charged.TransactionsCount,
			},
		}
	}

	if day.Spent != nil {
		mapped.Charged = TurnoverOrZero{
			t: &Turnover{
				Amount:            day.Spent.Amount,
				TransactionsCount: day.Spent.TransactionsCount,
			},
		}
	}
	return mapped
}
