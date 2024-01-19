package settings

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

func NewService(m map[string]any) Service {
	nm := map[string]any{
		"settings": m["parameters"].(map[string]any)["strichliste"],
	}
	return Service{m: nm}
}

type Service struct {
	m map[string]any
}

func (svc Service) Get(k string) any {
	parts := strings.Split(k, ".")

	var m map[string]any = svc.m["settings"].(map[string]any)
	for _, p := range parts[:len(parts)-1] {
		v := m[p]
		if v == nil {
			return nil
		}
		m = v.(map[string]any)
	}

	return m[parts[len(parts)-1]]
}

func (svc Service) GetInt(k string) (int, bool) {
	val := svc.Get(k)
	if val == nil {
		return 0, false
	}

	v, ok := val.(int)
	return v, ok
}

type Handler struct {
	settings Service
}

func NewHandler(svc Service) Handler {
	return Handler{settings: svc}
}

func (h Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, h.settings.m)
}
