package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	adomain "github.com/nerdbergev/strichliste-go/pkg/articles/domain"
	"github.com/nerdbergev/strichliste-go/pkg/barcodes"
	"github.com/nerdbergev/strichliste-go/pkg/barcodes/domain"
)

type Handler struct {
	svc barcodes.Service
}

func NewHandler(svc barcodes.Service) Handler {
	return Handler{
		svc: svc,
	}
}

func (h Handler) ListBarcodes(w http.ResponseWriter, r *http.Request) {
	barcodes, err := h.svc.GetAll()
	if err != nil {
		_ = render.Render(w, r, ErrServerError(err))
		return
	}
	if err := render.Render(w, r, NewBarcodeListResponse(barcodes)); err != nil {
		_ = render.Render(w, r, ErrServerError(err))
	}
}

func (h Handler) ListArticleBarcodes(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "aid")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidParamter("aid"))
		return
	}

	barcodes, err := h.svc.FindByArticleID(id)
	if err != nil {
		_ = render.Render(w, r, ErrServerError(err))
		return
	}
	if err := render.Render(w, r, NewBarcodeListResponse(barcodes)); err != nil {
		_ = render.Render(w, r, ErrServerError(err))
	}
}

func (h Handler) GetArticleBarcode(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "bid")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidParamter("bid"))
		return
	}
	barcode, err := h.svc.FindByID(id)
	if err != nil {
		h.renderError(w, r, err)
		return
	}
	_ = render.Render(w, r, NewBarcodeResponse(barcode))
}

type AddArticleBarcodeRequest struct {
	Barcode *string `json:"barcode"`
}

func (b *AddArticleBarcodeRequest) Bind(r *http.Request) error {
	if b.Barcode == nil {
		return ParameterMissingError{Name: "name"}
	}

	return nil
}

type ParameterMissingError struct {
	Name string
}

func (err ParameterMissingError) Error() string {
	return "" // we don't really need the Error function.
}

func (h Handler) AddArticleBarcode(w http.ResponseWriter, r *http.Request) {
	aidParam := chi.URLParam(r, "aid")
	aid, err := strconv.ParseInt(aidParam, 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidParamter("aid"))
		return
	}
	bReq := new(AddArticleBarcodeRequest)
	if err := render.Bind(r, bReq); err != nil {
		h.renderError(w, r, err)
		return
	}

	article, err := h.svc.AddArticleBarcode(aid, *bReq.Barcode)
	if err != nil {
		h.renderError(w, r, err)
		return
	}

	render.JSON(w, r, NewArticleResponse(article))
}

func (h Handler) DeleteArticleBarcode(w http.ResponseWriter, r *http.Request) {
	aidParam := chi.URLParam(r, "aid")
	aid, err := strconv.ParseInt(aidParam, 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidParamter("aid"))
		return
	}
	bidParam := chi.URLParam(r, "bid")
	bid, err := strconv.ParseInt(bidParam, 10, 64)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidParamter("bid"))
		return
	}

	article, err := h.svc.DeleteArticleBarcode(aid, bid)
	if err != nil {
		h.renderError(w, r, err)
		return
	}
	render.JSON(w, r, NewArticleResponse(article))
}

func (h Handler) renderError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		bnfErr domain.BarcodeNotFoundError
	)

	switch {
	case errors.As(err, &bnfErr):
		_ = render.Render(w, r, ErrBarcodeNotFound(bnfErr))
	default:
		_ = render.Render(w, r, ErrServerError(err))
	}
}

func NewBarcodeListResponse(barcodes []domain.Barcode) BarcodeListResponse {
	list := BarcodeListResponse{Barcodes: []Barcode{}, Count: len(barcodes)}
	for _, b := range barcodes {
		list.Barcodes = append(list.Barcodes, MapBarcode(b))
	}
	return list
}

type BarcodeListResponse struct {
	Barcodes []Barcode `json:"barcodes"`
	Count    int       `json:"count"`
}

func (br BarcodeListResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type Barcode struct {
	ID      int64     `json:"id"`
	Barcode string    `json:"barcode"`
	Created time.Time `json:"created"`
}

func MapBarcode(u domain.Barcode) Barcode {
	resp := Barcode{
		ID:      u.ID,
		Barcode: u.Barcode,
		Created: u.Created,
	}
	return resp
}

func NewBarcodeResponse(b domain.Barcode) BarcodeResponse {
	return BarcodeResponse{Barcode: MapBarcode(b)}
}

type BarcodeResponse struct {
	Barcode Barcode `json:"barcode"`
}

func (br BarcodeResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
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
		StatusCode: http.StatusInternalServerError,
		Error: Error{
			Message: "Internal Server Error",
			Code:    http.StatusInternalServerError,
		},
	}
}

func ErrInvalidParamter(name string) render.Renderer {
	return &ErrResponse{
		StatusCode: http.StatusBadRequest,
		Error: Error{
			Code:    http.StatusBadRequest,
			Class:   "App\\Exception\\ParameterInvalidException",
			Message: fmt.Sprintf("Parameter '%s' is invalid", name),
		},
	}
}

func ErrBarcodeNotFound(err domain.BarcodeNotFoundError) render.Renderer {
	return &ErrResponse{
		StatusCode: http.StatusNotFound,
		Error: Error{
			Class:   "App\\Exception\\BarcodeNotFoundException",
			Code:    http.StatusNotFound,
			Message: err.Error(),
		},
	}
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

type ArticleResponse struct {
	Article Article `json:"article"`
}

func (ar ArticleResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func NewArticleResponse(a adomain.Article) ArticleResponse {
	return ArticleResponse{Article: mapArticle(a)}
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
