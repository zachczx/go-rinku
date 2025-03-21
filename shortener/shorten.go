package shortener

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
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
	err := DB.QueryRow(`SELECT * FROM urls WHERE slug = $1`, slug).Scan(&rec.ID, &rec.Slug, &rec.Target, &rec.Hold, &rec.CreatedAt)
	if err != nil {
		return rec, fmt.Errorf("err: select slug: %w", err)
	}
	return rec, nil
}

func Log(URLID uuid.UUID, r *http.Request) error {
	_, err := DB.Exec(`INSERT INTO hits (url_id, referer, sec_ch_ua, sec_ch_ua_mobile, sec_ch_ua_platform, user_agent, created_at) VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		URLID,
		r.Header.Get("Referer"),
		r.Header.Get("Sec-Ch-Ua"),
		r.Header.Get("Sec-Ch-Ua-Mobile"),
		r.Header.Get("Sec-Ch-Ua-Platform"),
		r.Header.Get("User-Agent"))
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
