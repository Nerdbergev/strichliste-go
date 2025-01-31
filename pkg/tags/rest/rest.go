package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/tags"
	"github.com/nerdbergev/strichliste-go/pkg/tags/domain"
)

type Handler struct {
	svc tags.Service
}

func NewHandler(svc tags.Service) Handler {
	return Handler{
		svc: svc,
	}
}

func (h Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.svc.GetAll()
	if err != nil {
		h.renderError(w, r, err)
		return
	}
	if err := render.Render(w, r, NewTagListResponse(tags)); err != nil {
		_ = render.Render(w, r, ErrServerError(err))
	}
}

func (h Handler) ListArticleTags(w http.ResponseWriter, r *http.Request) {
	aid, err := strconv.ParseInt(chi.URLParam(r, "articleId"), 10, 64)
	if err != nil {
		h.renderError(w, r, err)
		return
	}
	tags, err := h.svc.GetArticleTags(aid)
	if err != nil {
		h.renderError(w, r, err)
		return
	}
	if err := render.Render(w, r, NewTagListResponse(tags)); err != nil {
		_ = render.Render(w, r, ErrServerError(err))
	}
}

func (h Handler) GetArticleTag(w http.ResponseWriter, r *http.Request) {
	tid, err := strconv.ParseInt(chi.URLParam(r, "tagId"), 10, 64)
	if err != nil {
		h.renderError(w, r, err)
		return
	}

	tag, err := h.svc.FindByID(tid)
	if err != nil {
		h.renderError(w, r, err)
		return
	}

	if err := render.Render(w, r, NewTagResponse(tag)); err != nil {
		_ = render.Render(w, r, ErrServerError(err))
		return
	}
}

func (h Handler) AddArticleTag(w http.ResponseWriter, r *http.Request) {
	aid, err := strconv.ParseInt(chi.URLParam(r, "articleId"), 10, 64)
	if err != nil {
		h.renderError(w, r, err)
		return
	}

	tReq := &TagRequest{}
	if err := render.Bind(r, tReq); err != nil {
		_ = render.Render(w, r, ErrServerError(err))
		return
	}
	article, err := h.svc.AddArticleTag(aid, tReq.TagParam)
	if err != nil {
		h.renderError(w, r, err)
		return
	}

	if err := render.Render(w, r, NewArticleResponse(article)); err != nil {
		_ = render.Render(w, r, ErrServerError(err))
		return
	}
}

func (h Handler) renderError(w http.ResponseWriter, r *http.Request, err error) {
	// var (
	// 	unfErr domain.UserNotFoundError
	// 	aeErr  domain.UserAlreadyExistsError
	// 	pmErr  ParameterMissingError
	// 	piErr  ParameterInvalidError
	// )

	switch {
	// case errors.As(err, &unfErr):
	// 	_ = render.Render(w, r, ErrUserNotFound(unfErr))
	// case errors.As(err, &aeErr):
	// 	_ = render.Render(w, r, ErrUserAlreadyExists(aeErr))
	// case errors.As(err, &pmErr):
	// 	_ = render.Render(w, r, ErrMissingParameter(pmErr.Name))
	// case errors.As(err, &piErr):
	// 	_ = render.Render(w, r, ErrInvalidParamter(piErr.Name))
	default:
		_ = render.Render(w, r, ErrServerError(err))
	}
}

type TagRequest struct {
	TagParam string `json:"tag"`
}

func (a *TagRequest) Bind(r *http.Request) error {
	// TODO: return proper errors
	// if a.NameParam == nil || *a.NameParam == "" {
	// 	return errors.New("missing required Article fields.")
	// }
	//
	// *a.NameParam = strings.TrimSpace(*a.NameParam)
	//
	// if a.AmountParam == nil {
	// 	return errors.New("missing required Article fields.")
	// }
	return nil
}

func NewTagListResponse(tags []domain.Tag) TagListResponse {
	list := TagListResponse{Count: len(tags), Tags: make([]Tag, 0, len(tags))}
	for _, t := range tags {
		list.Tags = append(list.Tags, Tag{
			ID:         t.ID,
			Tag:        t.Tag,
			Created:    t.Created,
			UsageCount: t.UsageCount,
		})
	}
	return list
}

func NewTagResponse(t domain.Tag) TagResponse {
	return TagResponse{Tag: Tag{
		ID:         t.ID,
		Tag:        t.Tag,
		Created:    t.Created,
		UsageCount: t.UsageCount,
	}}
}

type TagListResponse struct {
	Tags  []Tag `json:"tags"`
	Count int   `json:"count"`
}

func (tr TagListResponse) Render(w http.ResponseWriter, r *http.Request) error { return nil }

type TagResponse struct {
	Tag Tag `json:"tag"`
}

func (tr TagResponse) Render(w http.ResponseWriter, r *http.Request) error { return nil }

type Tag struct {
	ID         int64     `json:"id"`
	Tag        string    `json:"tag"`
	Created    time.Time `json:"created"`
	UsageCount int64     `json:"usageCount"`
}

type Error struct {
	Class   string `json:"class"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrResponse struct {
	StatusCode int   `json:"-"`
	Error      Error `json:"error"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.StatusCode)
	return nil
}

func ErrServerError(err error) render.Renderer {
	return &ErrResponse{
		Error: Error{
			Message: "Internal Server Error",
			Code:    http.StatusInternalServerError,
		},
	}
}

type ArticleResponse struct {
	Article Article `json:"article"`
}

func (ar ArticleResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewArticleResponse(a adomain.Article) ArticleResponse {
	return ArticleResponse{Article: mapArticle(a)}
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

func mapBarcodes(barcodes []adomain.Barcode) []Barcode {
	mapped := make([]Barcode, 0, len(barcodes))
	for _, b := range barcodes {
		mapped = append(mapped, Barcode{
			ID:      b.ID,
			Barcode: b.Barcode,
			Created: b.Created,
		})
	}
	return mapped
}

func mapArticle(a adomain.Article) Article {
	resp := Article{
		ID:         a.ID,
		Name:       a.Name,
		Amount:     a.Amount,
		IsActive:   a.IsActive,
		UsageCount: a.UsageCount,
		Created:    a.Created,
		Barcodes:   mapBarcodes(a.Barcodes),
	}

	if a.Precursor != nil {
		resp.Precursor = new(Article)
		*resp.Precursor = mapArticle(*a.Precursor)
	}
	return resp
}
