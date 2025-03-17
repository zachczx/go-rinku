package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"gorinku/shortener"
	"gorinku/templates"

	"github.com/a-h/templ"
	_ "github.com/jackc/pgx/v5/stdlib" // Pg driver
	"github.com/jmoiron/sqlx"
)

func main() {
	var err error
	pg := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	shortener.DB, err = sqlx.Open("pgx", pg)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})
	mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ADMIN HANDLER START:", r.URL.Path)
		records, err := shortener.ListAll()
		fmt.Println("INVOKER 1")
		if err != nil {
			fmt.Println(err)
			TemplRender(w, templates.Error(err.Error()))
			return
		}
		TemplRender(w, templates.AdminMain(records))
		fmt.Println("ADMIN HANDLER END")
	})
	mux.Handle("GET /admin/new", http.RedirectHandler("/admin", http.StatusSeeOther))
	mux.HandleFunc("POST /admin/new", func(w http.ResponseWriter, r *http.Request) {
		var hold bool = false
		fmt.Println(r.FormValue("hold"))
		if r.FormValue("hold") == "true" {
			hold = true
		}
		input := shortener.Record{Slug: r.FormValue("slug"), Target: r.FormValue("target"), Hold: hold}
		if err := shortener.Insert(input); err != nil {
			TemplRender(w, templates.Error(err.Error()))
		}
		records, err := shortener.ListAll()
		if err != nil {
			TemplRender(w, templates.Error(err.Error()))
			return
		}
		TemplRender(w, templates.AdminMain(records))
	})
	mux.HandleFunc("GET /admin/populate", func(w http.ResponseWriter, r *http.Request) {
		if err := shortener.Create(); err != nil {
			http.Error(w, "Error!", 500)
		}
		w.Write([]byte("written!"))
	})
	mux.HandleFunc("GET /{slug}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("SLUG HANDLER START:", r.URL.Path)
		slug := r.PathValue("slug")
		record, err := shortener.Check(slug)
		if err != nil {
			TemplRender(w, templates.Error(err.Error()))
			return
		}
		if !record.Hold {
			http.Redirect(w, r, record.Target, http.StatusSeeOther)
			return
		}
		TemplRender(w, templates.Holding(record.Target))
		fmt.Println("SLUG HANDLER END")
	})
	mux.Handle("GET /assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))

	http.ListenAndServe(":8001", mux)
}

func TemplRender(w http.ResponseWriter, c templ.Component) error {
	if err := c.Render(context.Background(), w); err != nil {
		return fmt.Errorf("error: templ render: %w", err)
	}
	return nil
}
