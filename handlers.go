package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"

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

// Library to grab IP addresses copied from https://github.com/vikram1565/request-ip.
// Standard headers list
var requestHeaders = []string{"X-Client-Ip", "X-Forwarded-For", "Cf-Connecting-Ip", "Fastly-Client-Ip", "True-Client-Ip", "X-Real-Ip", "X-Cluster-Client-Ip", "X-Forwarded", "Forwarded-For", "Forwarded"}

// GetClientIP - returns IP address string; The IP address if known, defaulting to empty string if unknown.
func GetClientIP(r *http.Request) string {
	for _, header := range requestHeaders {
		switch header {
		case "X-Forwarded-For": // Load-balancers (AWS ELB) or proxies.
			if host, correctIP := getClientIPFromXForwardedFor(r.Header.Get(header)); correctIP {
				return host
			}
		default:
			if host := r.Header.Get(header); isCorrectIP(host) {
				return host
			}
		}
	}

	//  remote address checks.
	host, _, splitHostPortError := net.SplitHostPort(r.RemoteAddr)
	if splitHostPortError == nil && isCorrectIP(host) {
		return host
	}
	return ""
}

// getClientIPFromXForwardedFor  - returns first known ip address else return empty string
func getClientIPFromXForwardedFor(headers string) (string, bool) {
	if headers == "" {
		return "", false
	}
	// x-forwarded-for may return multiple IP addresses in the format: "client IP, proxy 1 IP, proxy 2 IP"
	// Therefore, the right-most IP address is the IP address of the most recent proxy
	// and the left-most IP address is the IP address of the originating client.
	forwardedIPs := strings.Split(headers, ",")
	for _, IP := range forwardedIPs {
		// header can contain spaces too, strip those out.
		IP = strings.TrimSpace(IP)
		// make sure we only use this if it's ipv4 (ip:port)
		if split := strings.Split(IP, ":"); len(split) == 2 {
			IP = split[0]
		}
		if isCorrectIP(IP) {
			return IP, true
		}
	}
	return "", false
}

// isCorrectIP - return true if ip string is valid textual representation of an IP address, else returns false
func isCorrectIP(IP string) bool {
	return net.ParseIP(IP) != nil
}
