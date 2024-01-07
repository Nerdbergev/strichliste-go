package transactions

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/nerdbergev/shoppinglist-go/pkg/transactions/model"
)

type Handler struct {
	tr model.TransactionRepository
}

func NewHandler(tr model.TransactionRepository) Handler {
	return Handler{
		tr: tr,
	}
}

func (h Handler) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	uidParam := chi.URLParam(r, "id")
	uid, err := strconv.ParseInt(uidParam, 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	ts, err := h.tr.GetFromUser(uid)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	_ = render.Render(w, r, NewTransactionListResponse(ts))
}

func (h Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	uidParam := chi.URLParam(r, "id")
	uid, err := strconv.ParseInt(uidParam, 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	data := &TransactionRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	tr, err := h.tr.Deposit(uid, data.Amount)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	_ = render.Render(w, r, NewTransactionResponse(tr))

}

func NewTransactionResponse(t model.Transaction) TransactionResponse {
	return TransactionResponse{Transaction: MapTransaction(t)}
}

type TransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

func (tr TransactionResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewTransactionListResponse(ts []model.Transaction) TransactionListResponse {
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

func MapTransaction(t model.Transaction) Transaction {
	resp := Transaction{
		ID:        t.ID,
		User:      MapUser(t.User),
		Comment:   t.Comment.String,
		Amount:    t.Amount,
		IsDeleted: t.IsDeleted,
		Created:   t.Created.Format("2006-01-02 15:04:05"),
	}

	if t.ArticleID.Valid {
		resp.ArticleID = &t.ArticleID.Int64
	}
	if t.RecipientID.Valid {
		resp.RecipientID = &t.RecipientID.Int64
	}
	if t.SenderID.Valid {
		resp.SenderID = &t.SenderID.Int64
	}
	if t.Quantity.Valid {
		resp.Quantity = &t.Quantity.Int64
	}
	return resp
}

func MapUser(u model.User) User {
	resp := User{
		ID:         u.ID,
		Name:       u.Name,
		Email:      u.Email,
		Balance:    u.Balance,
		IsDisabled: u.Disabled,
		Created:    u.Created,
		Updated:    u.Updated,
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
	Balance    int        `json:"balance"`
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
	Amount int64 `json:"amount"`
}

func (u TransactionRequest) Bind(r *http.Request) error {
	return nil
}
