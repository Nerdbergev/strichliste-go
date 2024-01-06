package settings

import (
	"net/http"

	"github.com/go-chi/render"
)

func NewModel(m map[string]any) Model {
	nm := map[string]any{
		"settings": m["parameters"].(map[string]any)["strichliste"],
	}
	return Model{m: nm}
}

type Model struct {
	m map[string]any
}

func (sm Model) Get(k string) any {
	return sm.m[k]
}

type Handler struct {
	settings Model
}

func NewHandler(sm Model) Handler {
	return Handler{settings: sm}
}

func (h Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, h.settings.m)
}
