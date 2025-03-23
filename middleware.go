package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

type StatusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *StatusRecorder) WriteHeader(statusCode int) {
	rec.ResponseWriter.WriteHeader(statusCode)
	rec.status = statusCode
}

var limit int = 300 // setting limit for 300ms as minimum acceptable page load speed

func StatusLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := StatusRecorder{w, http.StatusOK}
		next.ServeHTTP(&rec, r)
		since := time.Since(start)

		if !strings.Contains(r.URL.Path, "/assets/") {
			var status, method, url, duration, encoding string

			switch rec.status {
			case http.StatusOK:
				status = pterm.Green(strconv.Itoa(rec.status))
			default:
				status = pterm.Red(strconv.Itoa(rec.status))
			}

			switch r.Method {
			case "GET":
				method = pterm.Green(r.Method)
			case "POST":
				method = pterm.Blue(r.Method)
			default:
				method = r.Method
			}

			switch r.URL.String() {
			case "":
				url = pterm.Red(r.URL)
			default:
				url = r.URL.String()
			}

			switch {
			case since < (time.Millisecond * time.Duration(limit)):
				duration = pterm.Green(since)
			default:
				duration = pterm.Red(since)
			}

			switch w.Header().Get("Content-Encoding") {
			case "br":
				encoding = pterm.LightWhite(w.Header().Get("Content-Encoding"))
			default:
				encoding = pterm.Red(w.Header().Get("Content-Encoding"))
			}
			fmt.Println(" ")
			pterm.DefaultSection.Println("Request!")
			pterm.Printf("[%v]-[%v]-[%v]-[%v]-[%v]\r\n\r\n", status, method, encoding, duration, url)
			fmt.Println("###################")
		}
	})
}
