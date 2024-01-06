package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/nerdbergev/shoppinglist-go/pkg/settings"
	"github.com/nerdbergev/shoppinglist-go/pkg/user"
	"github.com/nerdbergev/shoppinglist-go/pkg/user/model"
	"gopkg.in/yaml.v3"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	r := chi.NewRouter()

	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		log.Fatal(err)
	}

	b, err := os.ReadFile("strichliste.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var yml map[string]any
	err = yaml.Unmarshal(b, &yml)
	if err != nil {
		log.Fatal(err)
	}

	sm := settings.NewModel(yml)
	sh := settings.NewHandler(sm)

	um := model.New(db)

	uh := user.NewHandler(um)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Route("/api", func(r chi.Router) {
		r.Get("/user", uh.GetUsers)
		r.Post("/user", uh.CreateUser)
		r.Get("/settings", sh.GetSettings)
	})

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err.Error())
	}
}
