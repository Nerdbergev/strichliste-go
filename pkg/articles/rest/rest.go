package rest

import (
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/render"
	"github.com/nerdbergev/shoppinglist-go/pkg/articles"
	"github.com/nerdbergev/shoppinglist-go/pkg/articles/domain"
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
		articles, err = h.svc.GetAll()
		if err != nil {
			_ = render.Render(w, r, ErrRender(err))
			return
		}
		// } else {
		// 	isActive, perr := strconv.ParseBool(activeParam)
		// 	if perr != nil {
		// 		_ = render.Render(w, r, ErrRender(perr))
		// 		return
		// 	}
		// 	users, err = h.svc.FindByState(getAllRequest{active: isActive})
	}
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
	list := ArticleListResponse{Articles: []Article{}}
	for _, a := range articles {
		list.Articles = append(list.Articles, MapArticle(a))
	}
	return list
}

type ArticleListResponse struct {
	Count    int64     `json:"count"`
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
	return Article{}
}
