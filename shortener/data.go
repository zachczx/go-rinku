package shortener

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

type Record struct {
	ID        uuid.UUID
	Slug      string
	Target    string
	Hold      bool
	CreatedAt time.Time
}

func (rec Record) HoldString() string {
	if rec.Hold {
		return "true"
	}
	return "false"
}

func (rec Record) CreatedAtFormatted() string {
	return rec.CreatedAt.Format(time.RFC822)
}

var slugLength = 4

var chars = [62]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

func Insert(rec Record) error {
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

	rec.Target = HttpPrefix(rec.Target)

	_, err := DB.Exec(`INSERT INTO records (slug, target, hold, created_at) VALUES ($1, $2, $3, NOW())`, rec.Slug, rec.Target, rec.Hold)
	if err != nil {
		return fmt.Errorf("err: insert: %w", err)
	}
	return nil
}

func Create() error {
	_, err := DB.Exec(`CREATE TABLE records (record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), slug VARCHAR(255), target VARCHAR(2000), hold BOOLEAN, created_at TIMESTAMPTZ)`)
	if err != nil {
		return fmt.Errorf("err: create: %w", err)
	}
	return nil
}

func ListAll() ([]Record, error) {
	var records []Record
	var rec Record
	rows, err := DB.Query(`SELECT * FROM records`)
	if err != nil {
		return nil, fmt.Errorf("err: select: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&rec.ID, &rec.Slug, &rec.Target, &rec.Hold, &rec.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("err: scan: %w", err)
		}
		records = append(records, rec)
	}
	return records, nil
}

func Check(slug string) (Record, error) {
	var rec Record
	err := DB.QueryRow(`SELECT * FROM records WHERE slug = $1`, slug).Scan(&rec.ID, &rec.Slug, &rec.Target, &rec.Hold, &rec.CreatedAt)
	if err != nil {
		return rec, fmt.Errorf("err: select slug: %w", err)
	}
	return rec, nil
}

var urlRegex = regexp.MustCompile(`^(http:\/\/|https:\/\/).+`)

func HttpPrefix(target string) string {
	if !urlRegex.MatchString(target) {
		target = "http://" + target
	}
	return target
}
