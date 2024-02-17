package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/nerdbergev/strichliste-go/pkg/articles"
	arepo "github.com/nerdbergev/strichliste-go/pkg/articles/repository"
	arest "github.com/nerdbergev/strichliste-go/pkg/articles/rest"
	"github.com/nerdbergev/strichliste-go/pkg/settings"
	"github.com/nerdbergev/strichliste-go/pkg/transactions"
	trepo "github.com/nerdbergev/strichliste-go/pkg/transactions/repository"
	trest "github.com/nerdbergev/strichliste-go/pkg/transactions/rest"
	"github.com/nerdbergev/strichliste-go/pkg/users"
	urepo "github.com/nerdbergev/strichliste-go/pkg/users/repository"
	urest "github.com/nerdbergev/strichliste-go/pkg/users/rest"
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

	ss := settings.NewService(yml)
	sh := settings.NewHandler(ss)

	ar := arepo.New(db)
	asvc := articles.NewService(ar)
	ah := arest.NewHandler(asvc)

	ur := urepo.New(db)
	usvc, err := users.NewService(ss, ur)
	if err != nil {
		log.Fatal(err)
	}
	uh := urest.NewHandler(usvc)

	tr := trepo.New(db)
	tsvc := transactions.NewService(tr, ur, ar, ss)
	th := trest.NewHandler(tsvc)

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Get("/", uh.GetAll)
			r.Get("/{uid}", uh.FindById)
			r.Post("/{uid}", uh.UpdateUser)
			r.Route("/{uid}/transaction", func(r chi.Router) {
				r.Get("/", th.GetUserTransactions)
				r.Post("/", th.CreateTransaction)
				r.Delete("/{tid}", th.DeleteTransaction)
			})
			r.Post("/", uh.CreateUser)
		})
		r.Get("/settings", sh.GetSettings)
		r.Route("/article", func(r chi.Router) {
			r.Get("/", ah.List)
			r.Post("/", ah.CreateArticle)
			r.Post("/{aid}", ah.UpdateArticle)
			r.Delete("/{aid}", ah.DeactivateArticle)
		})
	})

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err.Error())
	}
}
