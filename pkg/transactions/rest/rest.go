package rest

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/nerdbergev/shoppinglist-go/pkg/transactions"
	"github.com/nerdbergev/shoppinglist-go/pkg/transactions/domain"
	udomain "github.com/nerdbergev/shoppinglist-go/pkg/users/domain"
)

type Handler struct {
	svc transactions.Service
}

func NewHandler(svc transactions.Service) Handler {
	return Handler{
		svc: svc,
	}
}

func (h Handler) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	uidParam := chi.URLParam(r, "id")
	uid, err := strconv.ParseInt(uidParam, 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	comment := r.URL.Query().Get("comment")
	if len(comment) > 255 {
		_ = render.Render(w, r, ErrRender(errors.New("comment invalid")))
		return
	}

	ts, err := h.svc.GetFromUser(uid)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	_ = render.Render(w, r, NewTransactionListResponse(ts))
}

func (h Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	uid, err := parseInt64(chi.URLParam(r, "id"))
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	data := &TransactionRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	tr, err := h.svc.ProcessTransaction(uid, data.Amount, nil, nil, nil, nil)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	_ = render.Render(w, r, NewTransactionResponse(tr))
}

func NewTransactionResponse(t domain.Transaction) TransactionResponse {
	return TransactionResponse{Transaction: MapTransaction(t)}
}

type TransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

func (tr TransactionResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewTransactionListResponse(ts []domain.Transaction) TransactionListResponse {
	list := TransactionListResponse{Transactions: []Transaction{}}
	for _, t := range ts {
		list.Transactions = append(list.Transactions, MapTransaction(t))
	}
	list.Count = len(ts)
	return list
}

type TransactionListResponse struct {
	Count        int           `json:"count"`
	Transactions []Transaction `json:"transactions"`
}

func (tr TransactionListResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func MapTransaction(t domain.Transaction) Transaction {

	resp := Transaction{
		ID:        t.ID,
		User:      MapUser(t.User),
		Amount:    t.Amount,
		IsDeleted: t.IsDeleted,
		Created:   t.Created.Format("2006-01-02 15:04:05"),
	}
	if t.Comment != nil {
		resp.Comment = *t.Comment
	}

	if t.Article != nil {
		resp.ArticleID = &t.Article.ID
	}
	if t.Recipient != nil {
		resp.RecipientID = &t.Recipient.ID
	}
	if t.Sender != nil {
		resp.SenderID = &t.Sender.ID
	}
	if t.Quantity != nil {
		resp.Quantity = t.Quantity
	}
	return resp
}

func MapUser(u udomain.User) User {
	resp := User{
		ID:      u.ID,
		Name:    u.Name,
		Balance: u.Balance,
		// IsDisabled: u.Disabled,
		Created: u.Created,
		Updated: u.Updated,
	}

	if u.Email != nil {
		resp.Email = *u.Email
	}
	return resp
}

type Transaction struct {
	ID           int64  `json:"id"`
	User         User   `json:"user"`
	ArticleID    *int64 `json:"article_id"`
	RecipientID  *int64 `json:"recipient"`
	SenderID     *int64 `json:"sender"`
	Quantity     *int64 `json:"quantity"`
	Comment      string `json:"comment"`
	Amount       int64  `json:"amount"`
	IsDeleted    bool   `json:"isDeleted"`
	IsDeleteable bool   `json:"isDeletable"`
	Created      string `json:"created"`
}

type User struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Balance    int64      `json:"balance"`
	IsActive   bool       `json:"isActive"`
	IsDisabled bool       `json:"isDisabled"`
	Created    time.Time  `json:"created"`
	Updated    *time.Time `json:"updated"`
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

type TransactionRequest struct {
	Amount      int64  `json:"amount"`
	Quantity    *int64 `json:"quantity"`
	Comment     string `json:"comment"`
	RecipientID *int64 `json:"recipientId"`
	ArticleID   *int64 `json:"articleId"`
}

func (u TransactionRequest) Bind(r *http.Request) error {
	if len(u.Comment) > 255 {
		return errors.New("invalid comment")
	}
	return nil
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
