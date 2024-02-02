package rest

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/nerdbergev/strichliste-go/pkg/articles"
	"github.com/nerdbergev/strichliste-go/pkg/articles/domain"
)

type Handler struct {
	svc articles.Service
}

func NewHandler(svc articles.Service) Handler {
	return Handler{
		svc: svc,
	}
}

func (h Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	activeParam := r.URL.Query().Get("active")
	var (
		articles []domain.Article
		err      error
	)
	if activeParam == "" {
		activeParam = "true"
	}
	isActive, perr := strconv.ParseBool(activeParam)
	if perr != nil {
		_ = render.Render(w, r, ErrRender(perr))
		return
	}
	articles, err = h.svc.GetAll(isActive)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Name < articles[j].Name
	})

	if err := render.Render(w, r, NewArticleListResponse(articles)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
	}
}

func (h Handler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	aReq := &ArticleRequest{}
	if err := render.Bind(r, aReq); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	created, err := h.svc.CreateArticle(aReq)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
	}
	render.JSON(w, r, NewArticleResponse(created))

}

type ArticleRequest struct {
	NameParam     string  `json:"name"`
	BarcodeParam  *string `json:"barcode"`
	IsActiveParam bool    `json:"isActive"`
	AmountParam   int64   `json:"amount"`
	// PrecursorParam *ArticleRequest `json:"precursor"` TODO
}

func (a ArticleRequest) Bind(r *http.Request) error {
	if a.NameParam == "" {
		return errors.New("missing required Article fields.")
	}
	return nil
}

func (a ArticleRequest) Name() string {
	return a.NameParam
}

func (a ArticleRequest) HasBarcode() bool {
	return a.BarcodeParam != nil
}

func (a ArticleRequest) Barcode() string {
	return *a.BarcodeParam
}

func (a ArticleRequest) IsActive() bool {
	return a.IsActiveParam
}

func (a ArticleRequest) Amount() int64 {
	return a.AmountParam
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

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

func NewArticleListResponse(articles []domain.Article) ArticleListResponse {
	list := ArticleListResponse{Count: len(articles), Articles: []Article{}}
	for _, a := range articles {
		list.Articles = append(list.Articles, MapArticle(a))
	}
	return list
}

type ArticleListResponse struct {
	Count    int       `json:"count"`
	Articles []Article `json:"articles"`
}

func (ar ArticleListResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type Article struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Barcode    *string   `json:"barcode"`
	Amount     int64     `json:"amount"`
	IsActive   bool      `json:"isActive"`
	UsageCount int64     `json:"usageCount"`
	Precursor  *Article  `json:"precursor"`
	Created    time.Time `json:"created"`
}

func MapArticle(a domain.Article) Article {
	resp := Article{
		ID:         a.ID,
		Name:       a.Name,
		Amount:     a.Amount,
		IsActive:   a.IsActive,
		UsageCount: a.UsageCount,
		Created:    a.Created,
	}

	if a.Barcode != nil {
		resp.Barcode = new(string)
		*resp.Barcode = *a.Barcode
	}
	return resp
}

type ArticleResponse struct {
	Article Article `json:"article"`
}

func (ar ArticleResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewArticleResponse(a domain.Article) ArticleResponse {
	return ArticleResponse{Article: MapArticle(a)}
}
