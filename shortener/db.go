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
	{name: "urls", query: `DROP TABLE IF EXISTS urls`},
	{name: "records", query: `DROP TABLE IF EXISTS records`},
	{name: "hits", query: `DROP TABLE IF EXISTS hits`},
}

var create = []Query{
	{name: "urls", query: `CREATE TABLE urls (url_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), slug VARCHAR(255), target VARCHAR(2000), hold BOOLEAN, created_at TIMESTAMPTZ)`},
	{name: "hits", query: `CREATE TABLE hits (hits_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), url_id UUID REFERENCES urls(url_id) ON DELETE CASCADE, referer VARCHAR(2000), created_at TIMESTAMPTZ)`},
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
