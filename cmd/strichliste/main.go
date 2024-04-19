package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/chi-middleware/proxy"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/nerdbergev/strichliste-go/pkg/articles"
	arepo "github.com/nerdbergev/strichliste-go/pkg/articles/repository"
	arest "github.com/nerdbergev/strichliste-go/pkg/articles/rest"
	"github.com/nerdbergev/strichliste-go/pkg/metrics"
	mrepo "github.com/nerdbergev/strichliste-go/pkg/metrics/repository"
	mrest "github.com/nerdbergev/strichliste-go/pkg/metrics/rest"
	"github.com/nerdbergev/strichliste-go/pkg/settings"
	"github.com/nerdbergev/strichliste-go/pkg/transactions"
	trepo "github.com/nerdbergev/strichliste-go/pkg/transactions/repository"
	trest "github.com/nerdbergev/strichliste-go/pkg/transactions/rest"
	"github.com/nerdbergev/strichliste-go/pkg/users"
	urepo "github.com/nerdbergev/strichliste-go/pkg/users/repository"
	urest "github.com/nerdbergev/strichliste-go/pkg/users/rest"
	"gopkg.in/yaml.v3"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xo/dburl"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	if err := run(context.Background(), os.Stdout, os.Args, os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string, getenv func(string) string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	dbUrl := getenv("DATABASE_URL")
	if strings.TrimSpace(dbUrl) == "" {
		return errors.New(".env is missing DATABASE_URL")
	}
	db, err := dburl.Open(dbUrl)
	if err != nil {
		return err
	}

	b, err := os.ReadFile("strichliste.yaml")
	if err != nil {
		return err
	}

	var yml map[string]any
	err = yaml.Unmarshal(b, &yml)
	if err != nil {
		return err
	}

	ss := settings.NewService(yml)
	settingsHandler := settings.NewHandler(ss)

	ar := arepo.New(db)
	asvc := articles.NewService(ar)
	articlesHandler := arest.NewHandler(asvc)

	ur := urepo.New(db)
	usvc, err := users.NewService(ss, ur)
	if err != nil {
		return err
	}
	usersHandler := urest.NewHandler(usvc)

	tr := trepo.New(db)
	tsvc := transactions.NewService(tr, ur, ar, ss)
	transactionsHandler := trest.NewHandler(tsvc)

	mr := mrepo.New(db)
	msvc := metrics.NewService(mr)
	metricsHandler := mrest.NewHandler(msvc)

	trustedProxies := strings.Split(getenv("TRUSTED_PROXIES"), ",")
	srv := NewServer(
		settingsHandler,
		articlesHandler,
		usersHandler,
		transactionsHandler,
		metricsHandler,
		trustedProxies,
	)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort("127.0.0.1", "8081"),
		Handler: srv,
	}

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		fmt.Println("shutdown received")
		defer cancel()
	}()
	wg.Wait()
	return nil
}

func NewServer(
	sh settings.Handler,
	ah arest.Handler,
	uh urest.Handler,
	th trest.Handler,
	mh mrest.Handler,
	trustedProxies []string,
) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	for _, addr := range trustedProxies {
		r.Use(proxy.ForwardedHeaders(
			proxy.NewForwardedHeadersOptions().
				ClearTrustedProxies().
				AddTrustedProxy(strings.TrimSpace(addr)),
		))
	}

	addRoutes(
		r,
		sh,
		ah,
		uh,
		th,
		mh,
	)

	return r
}

func addRoutes(
	router chi.Router,
	sh settings.Handler,
	ah arest.Handler,
	uh urest.Handler,
	th trest.Handler,
	mh mrest.Handler,
) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Route("/api", func(router chi.Router) {
		router.Route("/user", func(router chi.Router) {
			router.Get("/", uh.GetAll)
			router.Get("/{uid}", uh.FindById)
			router.Post("/{uid}", uh.UpdateUser)
			router.Route("/{uid}/transaction", func(router chi.Router) {
				router.Get("/", th.GetUserTransactions)
				router.Post("/", th.CreateTransaction)
				router.Delete("/{tid}", th.DeleteTransaction)
			})
			router.Post("/", uh.CreateUser)
		})
		router.Get("/settings", sh.GetSettings)
		router.Route("/article", func(router chi.Router) {
			router.Get("/", ah.List)
			router.Post("/", ah.CreateArticle)
			router.Post("/{aid}", ah.UpdateArticle)
			router.Delete("/{aid}", ah.DeactivateArticle)
		})
		router.Route("/metrics", func(r chi.Router) {
			router.Get("/", mh.GetMetrics)
		})
	})
}
