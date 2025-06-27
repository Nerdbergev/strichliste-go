package rest

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	// TODO: pagination
	isActive, perr := parseActiveParam(r)
	if perr != nil {
		_ = render.Render(w, r, ErrRender(perr))
		return
	}

	barcode := strings.TrimSpace(r.URL.Query().Get("barcode"))

	precursor, perr := parsePrecursorParam(r)
	if perr != nil {
		_ = render.Render(w, r, ErrRender(perr))
		return
	}

	ancestorVal, ok, perr := parseAncestorParam(r)
	if perr != nil {
		_ = render.Render(w, r, ErrRender(perr))
		return
	}
	var ancestor *bool
	if ok {
		ancestor = &ancestorVal
	}

	var (
		articles []domain.Article
		err      error
	)

	articles, err = h.svc.GetAll(isActive, precursor, barcode, ancestor)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	count, err := h.svc.CountActive()
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	if err := render.Render(w, r, NewArticleListResponse(articles, count)); err != nil {
		_ = render.Render(w, r, ErrRender(err))
	}
}

func (h Handler) FindById(w http.ResponseWriter, r *http.Request) {
	aid, err := strconv.ParseInt(chi.URLParam(r, "aid"), 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	user, err := h.svc.FindById(domain.ArticleID(aid))
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	_ = render.Render(w, r, NewArticleResponse(user))
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
		return
	}
	render.JSON(w, r, NewArticleResponse(created))

}

func (h Handler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	aid, err := strconv.ParseInt(chi.URLParam(r, "aid"), 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	aReq := &ArticleRequest{}
	if err := render.Bind(r, aReq); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	updated, err := h.svc.UpdateArticle(domain.ArticleID(aid), aReq)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	render.JSON(w, r, NewArticleResponse(updated))
}

func (h Handler) DeactivateArticle(w http.ResponseWriter, r *http.Request) {
	aid, err := strconv.ParseInt(chi.URLParam(r, "aid"), 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}

	deleted, err := h.svc.DeactivateArticle(domain.ArticleID(aid))
	if err != nil {
		_ = render.Render(w, r, ErrRender(err))
		return
	}
	render.JSON(w, r, NewArticleResponse(deleted))
}

type ArticleRequest struct {
	NameParam     *string `json:"name"`
	BarcodeParam  *string `json:"barcode"`
	IsActiveParam bool    `json:"isActive"`
	AmountParam   *int64  `json:"amount"`
	// Not implemented at the moment since the original strichliste doesn't too.
	PrecursorParam *ArticleRequest `json:"precursor"`
}

func (a *ArticleRequest) Bind(r *http.Request) error {
	// TODO: return proper errors
	if a.NameParam == nil || *a.NameParam == "" {
		return errors.New("missing required Article fields.")
	}

	*a.NameParam = strings.TrimSpace(*a.NameParam)

	if a.AmountParam == nil {
		return errors.New("missing required Article fields.")
	}
	return nil
}

func (a ArticleRequest) Name() string {
	return *a.NameParam
}

func (a ArticleRequest) HasBarcode() bool {
	return a.BarcodeParam != nil && strings.TrimSpace(*a.BarcodeParam) != ""
}

func (a ArticleRequest) Barcode() string {
	return *a.BarcodeParam
}

func (a ArticleRequest) IsActive() bool {
	return a.IsActiveParam
}

func (a ArticleRequest) Amount() int64 {
	return *a.AmountParam
}

func (a ArticleRequest) HasPrecursor() bool {
	return a.PrecursorParam != nil
}

func (a ArticleRequest) Precursor() articles.ArticleRequest {
	return *a.PrecursorParam
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

func NewArticleListResponse(articles []domain.Article, count int) ArticleListResponse {
	list := ArticleListResponse{Count: count, Articles: []Article{}}
	for _, a := range articles {
		list.Articles = append(list.Articles, mapArticle(a))
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

// type Barcode struct {
// 	ID      int64     `json:"id"`
// 	Barcode string    `json:"barcode"`
// 	Created time.Time `json:"created"`
// }
//
// type Tag struct {
// }

type Article struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Barcodes   []Barcode `json:"barcodes"`
	Tags       []Tag     `json:"tags"`
	Amount     int64     `json:"amount"`
	IsActive   bool      `json:"isActive"`
	UsageCount int64     `json:"usageCount"`
	Precursor  *Article  `json:"precursor"`
	Created    time.Time `json:"created"`
}

// func mapBarcodes(barcodes []domain.Barcode) []Barcode {
// 	mapped := make([]Barcode, 0, len(barcodes))
// 	for _, b := range barcodes {
// 		mapped = append(mapped, Barcode{
// 			ID:      b.ID,
// 			Barcode: b.Barcode,
// 			Created: b.Created,
// 		})
// 	}
// 	return mapped
// }

func mapArticle(a domain.Article) Article {
	resp := Article{
		ID:         int64(a.ID),
		Name:       a.Name,
		Amount:     a.Amount,
		IsActive:   a.IsActive,
		UsageCount: a.UsageCount,
		Created:    a.Created,
		Barcodes:   mapBarcodes(a.Barcodes),
		Tags:       mapTags(a.Tags),
	}

	if a.Precursor != nil {
		resp.Precursor = new(Article)
		*resp.Precursor = mapArticle(*a.Precursor)
	}
	return resp
}

func mapTags(ts []domain.Tag) []Tag {
	mapped := make([]Tag, 0, len(ts))
	for _, t := range ts {
		mapped = append(mapped, Tag{
			ID:         int64(t.ID),
			Tag:        t.Tag,
			Created:    t.Created,
			UsageCount: t.UsageCount,
		})
	}

	return mapped
}

type ArticleResponse struct {
	Article Article `json:"article"`
}

func (ar ArticleResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewArticleResponse(a domain.Article) ArticleResponse {
	return ArticleResponse{Article: mapArticle(a)}
}

func parseActiveParam(r *http.Request) (isActive bool, err error) {
	activeParam := strings.TrimSpace(r.URL.Query().Get("active"))
	if activeParam == "" {
		return true, nil
	}
	return strconv.ParseBool(activeParam)
}

func parsePrecursorParam(r *http.Request) (precursor bool, err error) {
	precursorParam := strings.TrimSpace(r.URL.Query().Get("precursor"))
	if precursorParam == "" {
		return true, nil
	}
	return strconv.ParseBool(precursorParam)
}

// parseAncestorParam parses the ancestor query param.
// It returns ok = false if the query param was not sent.
func parseAncestorParam(r *http.Request) (ancestor, ok bool, err error) {
	ancestorParam := strings.TrimSpace(r.URL.Query().Get("ancestor"))
	if ancestorParam == "" {
		return false, false, nil
	}
	parsed, err := strconv.ParseBool(ancestorParam)
	return parsed, true, err
}
