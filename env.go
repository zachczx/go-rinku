//go:build windows

package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Loaded env variables...")
}
