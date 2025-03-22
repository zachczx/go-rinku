package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"rinku/shortener"

	"github.com/a-h/templ"
	_ "github.com/jackc/pgx/v5/stdlib" // Pg driver
	"github.com/jmoiron/sqlx"
)

var emptyString string

var ctx = context.Background()

var domain = "zczx.org/"

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
	mux.HandleFunc("GET /{$}", landingHandler)
	mux.HandleFunc("GET /{slug}", shortenHandler)
	mux.Handle("GET /assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))

	// Admin routes
	mux.HandleFunc("GET /admin/login", adminLoginHandler)
	mux.Handle("GET /admin", service.RequireAuthentication(user, http.HandlerFunc(adminMainHandler)))
	mux.Handle("GET /admin/analyze/{ID}", service.RequireAuthentication(user, http.HandlerFunc(adminAnalyzeHandler)))
	mux.HandleFunc("GET /admin/authenticate", service.authenticateHandler)
	mux.Handle("GET /admin/new", http.RedirectHandler("/admin", http.StatusSeeOther))
	mux.Handle("POST /admin/new", service.RequireAuthentication(user, http.HandlerFunc(newURLHandler)))
	mux.Handle("POST /admin/delete/{ID}", service.RequireAuthentication(user, http.HandlerFunc(deleteURLHandler)))
	mux.Handle("POST /admin/login/sendlink", http.HandlerFunc(service.sendMagicLinkHandler))
	mux.Handle("GET /admin/reset", service.RequireAuthentication(user, http.HandlerFunc(resetDestroyHandler)))
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
