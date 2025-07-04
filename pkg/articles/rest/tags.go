package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/nerdbergev/strichliste-go/pkg/articles/domain"
)

func (h Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.svc.GetAllTags()
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
	tags, err := h.svc.GetArticleTags(domain.ArticleID(aid))
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

	tag, err := h.svc.FindTagByID(domain.TagID(tid))
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
	article, err := h.svc.AddArticleTag(domain.ArticleID(aid), tReq.TagParam)
	if err != nil {
		h.renderError(w, r, err)
		return
	}

	if err := render.Render(w, r, NewArticleResponse(article)); err != nil {
		_ = render.Render(w, r, ErrServerError(err))
		return
	}
}

func (h Handler) DeleteArticleTag(w http.ResponseWriter, r *http.Request) {
	aid, err := strconv.ParseInt(chi.URLParam(r, "articleId"), 10, 64)
	if err != nil {
		h.renderError(w, r, err)
		return
	}

	tid, err := strconv.ParseInt(chi.URLParam(r, "tagId"), 10, 64)
	if err != nil {
		h.renderError(w, r, err)
		return
	}
	article, err := h.svc.DeleteArticleTag(domain.ArticleID(aid), domain.TagID(tid))
	if err != nil {
		h.renderError(w, r, err)
		return
	}
	if err := render.Render(w, r, NewArticleResponse(article)); err != nil {
		_ = render.Render(w, r, ErrServerError(err))
		return
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
			ID:         int64(t.ID),
			Tag:        t.Tag,
			Created:    t.Created,
			UsageCount: t.UsageCount,
		})
	}
	return list
}

func NewTagResponse(t domain.Tag) TagResponse {
	return TagResponse{Tag: Tag{
		ID:         int64(t.ID),
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
