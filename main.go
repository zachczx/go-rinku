package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gorinku/shortener"
	"gorinku/templates"

	"github.com/a-h/templ"
	_ "github.com/jackc/pgx/v5/stdlib" // Pg driver
	"github.com/jmoiron/sqlx"
)

var emptyString string

var ctx = context.Background()

func main() {
	var err error
	pg := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	shortener.DB, err = sqlx.Open("pgx", pg)
	if err != nil {
		log.Fatal(err)
	}

	user := &User{Username: ""}

	service := NewAuthService(
		os.Getenv("STYTCH_PROJECT_ID"),
		os.Getenv("STYTCH_SECRET"),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.Landing())
	})

	mux.HandleFunc("GET /{slug}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("SLUG HANDLER START:", r.URL.Path)
		slug := r.PathValue("slug")
		record, err := shortener.Check(slug)
		if err != nil {
			TemplRender(w, r, templates.Error(emptyString))
			return
		}
		err = shortener.Log(record.ID, r)
		if err != nil {
			TemplRender(w, r, templates.Error(emptyString))
			return
		}
		if !record.Hold {
			http.Redirect(w, r, record.Target, http.StatusSeeOther)
			return
		}
		TemplRender(w, r, templates.Holding(record.Target))
		fmt.Println("SLUG HANDLER END")
	})
	mux.Handle("GET /assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))

	// Admin routes
	mux.HandleFunc("GET /admin/login", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.Login())
	})
	mux.Handle("GET /admin", service.RequireAuthentication(user, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ADMIN HANDLER START:", r.URL.Path)
		urls, err := shortener.ListAll()
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error(emptyString))
			return
		}
		TemplRender(w, r, templates.AdminMain(urls))
		fmt.Println("ADMIN HANDLER END")
	})))
	mux.HandleFunc("GET /admin/authenticate", service.authenticateHandler)
	mux.Handle("GET /admin/new", http.RedirectHandler("/admin", http.StatusSeeOther))
	mux.HandleFunc("POST /admin/new", newURLHandler)
	mux.HandleFunc("POST /admin/delete/{ID}", deleteURLHandler)
	mux.HandleFunc("POST /admin/login/sendlink", service.sendMagicLinkHandler)
	mux.HandleFunc("GET /admin/reset", resetDestroyHandler)
	mux.Handle("GET /admin/logout", service.logout(user))
	server := &http.Server{
		Addr:              os.Getenv("LISTEN_ADDR"),
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func TemplRender(w http.ResponseWriter, r *http.Request, c templ.Component) {
	if err := c.Render(r.Context(), w); err != nil {
		fmt.Println("error: templ render: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
