package shortener

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/pgtype"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

type URL struct {
	ID        uuid.UUID
	Slug      string
	Target    string
	Hold      bool
	CreatedAt time.Time
	Hits      sql.NullInt64
}

type Hit struct {
	ID              uuid.UUID
	URLID           uuid.UUID
	Referer         string
	SecChUa         string
	SecChUaMobile   string
	SecChUaPlatform string
	UserAgent       string
	IPAddr          pgtype.Inet
	CreatedAt       time.Time
}

func (url URL) HoldString() string {
	if url.Hold {
		return "true"
	}
	return "false"
}

func (url URL) HitsString() string {
	if !url.Hits.Valid {
		return "0"
	}
	return strconv.FormatInt(url.Hits.Int64, 10)
}

func (rec URL) CreatedAtFormatted() string {
	return rec.CreatedAt.Format("2 Jan 2006")
}

func (h Hit) CreatedAtFormatted() string {
	return h.CreatedAt.Format("2 Jan 2006")
}

func (h Hit) IPAddrString() string {
	if val, ok := h.IPAddr.Get().(*net.IPNet); ok {
		return val.String()
	}
	return ""
}

var slugLength = 4

var chars = [62]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

func Insert(rec URL) error {
	fmt.Println(rec)
	var randomSlug string

	if len(rec.Slug) == 0 {
		for i := 0; i < slugLength; i++ {
			idx := rand.IntN(len(chars) - 1)
			chosen := chars[idx]
			randomSlug += chosen
		}
		rec.Slug = randomSlug
	}

	rec.Target = HTTPPrefix(rec.Target)

	_, err := DB.Exec(`INSERT INTO urls (slug, target, hold, created_at) VALUES ($1, $2, $3, NOW())`, rec.Slug, rec.Target, rec.Hold)
	if err != nil {
		return fmt.Errorf("err: insert: %w", err)
	}
	return nil
}

func ListAll() ([]URL, error) {
	var urls []URL
	var rec URL
	rows, err := DB.Query(`SELECT urls.url_id, urls.slug, urls.target, urls.hold, urls.created_at, hits.cnt FROM urls
							LEFT JOIN (SELECT url_id, COUNT(1) as cnt FROM hits GROUP BY hits.url_id) as hits 
							ON hits.url_id = urls.url_id`)
	if err != nil {
		return nil, fmt.Errorf("err: select: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&rec.ID, &rec.Slug, &rec.Target, &rec.Hold, &rec.CreatedAt, &rec.Hits)
		if err != nil {
			return nil, fmt.Errorf("err: scan: %w", err)
		}
		urls = append(urls, rec)
	}
	return urls, nil
}

func Check(slug string) (URL, error) {
	var rec URL
	err := DB.QueryRow(`SELECT url_id, slug, target, hold, created_at FROM urls WHERE slug = $1`, slug).Scan(&rec.ID, &rec.Slug, &rec.Target, &rec.Hold, &rec.CreatedAt)
	if err != nil {
		return rec, fmt.Errorf("err: select slug: %w", err)
	}
	return rec, nil
}

func Log(URLID uuid.UUID, r *http.Request) error {
	IPAddr := GetClientIP(r)
	_, err := DB.Exec(`INSERT INTO hits (url_id, referer, sec_ch_ua, sec_ch_ua_mobile, sec_ch_ua_platform, user_agent, ip_address, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`,
		URLID,
		r.Header.Get("Referer"),
		r.Header.Get("Sec-Ch-Ua"),
		r.Header.Get("Sec-Ch-Ua-Mobile"),
		r.Header.Get("Sec-Ch-Ua-Platform"),
		r.Header.Get("User-Agent"),
		IPAddr,
	)
	if err != nil {
		return fmt.Errorf("err: select slug: %w", err)
	}
	return nil
}

var urlRegex = regexp.MustCompile(`^(http:\/\/|https:\/\/).+`)

func HTTPPrefix(target string) string {
	if !urlRegex.MatchString(target) {
		fmt.Println("HTTP/HTTPS not found in url prefix")
		target = "https://" + target
	}
	return target
}

func Delete(id uuid.UUID) error {
	_, err := DB.Exec(`DELETE FROM urls WHERE url_id = $1`, id)
	if err != nil {
		return fmt.Errorf("err: delete: %w", err)
	}
	return nil
}

func Analyze(ID uuid.UUID) ([]Hit, error) {
	var hits []Hit
	var hit Hit
	rows, err := DB.Query(`SELECT hit_id, url_id, referer, sec_ch_ua, sec_ch_ua_mobile, sec_ch_ua_platform, user_agent, ip_address, created_at FROM hits WHERE url_id = $1`, ID)
	if err != nil {
		return nil, fmt.Errorf("err: select: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&hit.ID, &hit.URLID, &hit.Referer, &hit.SecChUa, &hit.SecChUaMobile, &hit.SecChUaPlatform, &hit.UserAgent, &hit.IPAddr, &hit.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("err: scan: %w", err)
		}

		hits = append(hits, hit)
	}
	return hits, nil
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
