package shortener

import "fmt"

type Query struct {
	name  string
	query string
}

func Create() error {
	for _, v := range create {
		if _, err := DB.Exec(v.query); err != nil {
			return fmt.Errorf("err: drop: %v: - %w", v.name, err)
		}
	}
	return nil
}

var drop = []Query{
	{name: "hits", query: `DROP TABLE IF EXISTS hits`},
	{name: "urls", query: `DROP TABLE IF EXISTS urls`},
}

var create = []Query{
	{name: "urls", query: `CREATE TABLE urls (url_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), slug VARCHAR(255), target VARCHAR(2000), hold BOOLEAN, created_at TIMESTAMPTZ)`},
	{name: "hits", query: `CREATE TABLE hits (hit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), url_id UUID REFERENCES urls(url_id) ON DELETE CASCADE, referer VARCHAR(2000), sec_ch_ua VARCHAR(500), sec_ch_ua_mobile VARCHAR(500), sec_ch_ua_platform VARCHAR(500), user_agent VARCHAR(500), created_at TIMESTAMPTZ)`},
}

func Reset() error {
	for _, v := range drop {
		if _, err := DB.Exec(v.query); err != nil {
			return fmt.Errorf("err: drop: %v: - %w", v.name, err)
		}
	}

	if err := Create(); err != nil {
		return fmt.Errorf("err: reset: %w", err)
	}
	return nil
}
