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
	"github.com/nerdbergev/shoppinglist-go/pkg/transactions"
	trepo "github.com/nerdbergev/shoppinglist-go/pkg/transactions/repository"
	trest "github.com/nerdbergev/shoppinglist-go/pkg/transactions/rest"
	"github.com/nerdbergev/shoppinglist-go/pkg/users"
	urepo "github.com/nerdbergev/shoppinglist-go/pkg/users/repository"
	urest "github.com/nerdbergev/shoppinglist-go/pkg/users/rest"
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

	sm := settings.NewService(yml)
	sh := settings.NewHandler(sm)

	ur := urepo.New(db)
	usvc, err := users.NewService(sm, ur)
	if err != nil {
		log.Fatal(err)
	}
	uh := urest.NewHandler(usvc)

	tr := trepo.New(db)
	tsvc := transactions.NewService(tr)
	th := trest.NewHandler(tsvc)

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Get("/", uh.GetAll)
			r.Get("/{id}", uh.FindById)
			r.Route("/{id}/transaction", func(r chi.Router) {
				r.Get("/", th.GetUserTransactions)
				r.Post("/", th.CreateTransaction)
			})
			r.Post("/", uh.CreateUser)
		})
		r.Get("/settings", sh.GetSettings)
	})

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err.Error())
	}
}
