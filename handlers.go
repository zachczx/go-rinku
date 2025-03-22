package main

import (
	"fmt"
	"net/http"

	"rinku/shortener"
	"rinku/templates"

	"github.com/google/uuid"
)

func newURLHandler(w http.ResponseWriter, r *http.Request) {
	urlPrefix := r.FormValue("protocol")
	if urlPrefix != "http" && urlPrefix != "https" {
		http.Error(w, "Error!", 500)
		return
	}
	target := urlPrefix + "://" + r.FormValue("target")
	hold := false
	fmt.Println(r.FormValue("hold"))
	if r.FormValue("hold") == "true" {
		hold = true
	}
	input := shortener.URL{Slug: r.FormValue("slug"), Target: target, Hold: hold}
	if err := shortener.Insert(input); err != nil {
		fmt.Println(err)
		TemplRender(w, r, templates.Error(emptyString))
	}
	urls, err := shortener.ListAll()
	if err != nil {
		fmt.Println(err)
		TemplRender(w, r, templates.Error(emptyString))
		return
	}
	TemplRender(w, r, templates.AdminMain(urls, domain))
}

func resetDestroyHandler(w http.ResponseWriter, r *http.Request) {
	if err := shortener.Reset(); err != nil {
		fmt.Println(err)
		http.Error(w, "Error!", 500)
	}
	if _, err := w.Write([]byte("written!")); err != nil {
		fmt.Println(err)
		TemplRender(w, r, templates.Error(emptyString))
		return
	}
}

func deleteURLHandler(w http.ResponseWriter, r *http.Request) {
	ID := r.PathValue("ID")
	uuidID, err := uuid.Parse(ID)
	if err != nil {
		fmt.Println(err)
		TemplRender(w, r, templates.Error(emptyString))
		return
	}
	if err := shortener.Delete(uuidID); err != nil {
		fmt.Println(err)
		TemplRender(w, r, templates.Error(emptyString))
		return
	}
	urls, err := shortener.ListAll()
	if err != nil {
		fmt.Println(err)
		TemplRender(w, r, templates.Error(emptyString))
		return
	}
	TemplRender(w, r, templates.AdminMain(urls, domain))
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SLUG HANDLER START:", r.URL.Path)
	slug := r.PathValue("slug")
	record, err := shortener.Check(slug)
	if err != nil {
		fmt.Println(err)
		TemplRender(w, r, templates.Error(emptyString))
		return
	}
	err = shortener.Log(record.ID, r)
	if err != nil {
		fmt.Println(err)
		TemplRender(w, r, templates.Error(emptyString))
		return
	}
	if !record.Hold {
		http.Redirect(w, r, record.Target, http.StatusSeeOther)
		return
	}
	TemplRender(w, r, templates.Holding(record.Target))
	fmt.Println("SLUG HANDLER END")
}

func landingHandler(w http.ResponseWriter, r *http.Request) {
	TemplRender(w, r, templates.Landing())
}

func adminMainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ADMIN HANDLER START:", r.URL.Path)
	urls, err := shortener.ListAll()
	if err != nil {
		fmt.Println(err)
		TemplRender(w, r, templates.Error(emptyString))
		return
	}
	TemplRender(w, r, templates.AdminMain(urls, domain))
	fmt.Println("ADMIN HANDLER END")
}

func adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	TemplRender(w, r, templates.Login())
}

func adminAnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	ID := r.PathValue("ID")
	uuidID, err := uuid.Parse(ID)
	if err != nil {
		fmt.Println(err)
		TemplRender(w, r, templates.Error(emptyString))
		return
	}
	hits, err := shortener.Analyze(uuidID)
	if err != nil {
		fmt.Println(err)
		TemplRender(w, r, templates.Error(emptyString))
		return
	}

	TemplRender(w, r, templates.AdminAnalyze(hits))
}
