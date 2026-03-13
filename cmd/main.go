package main

import (
	"fmt"
	"log"

	"github.com/siwar/price-tag/internal/scraper"
)

func main() {
	s := scraper.New()

	prices, err := s.Scrape("https://example.com")
	if err != nil {
		log.Fatalf("scrape error: %v", err)
	}

	for _, p := range prices {
		fmt.Printf("Name: %s | Price: %.2f %s\n", p.Name, p.Amount, p.Currency)
	}
}
