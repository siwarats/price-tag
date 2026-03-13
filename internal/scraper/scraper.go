package scraper

import (
	"fmt"

	"github.com/siwar/price-tag/pkg/models"
)

// Scraper fetches and parses price data from web pages.
type Scraper struct{}

// New creates a new Scraper instance.
func New() *Scraper {
	return &Scraper{}
}

// Scrape fetches prices from the given URL.
func (s *Scraper) Scrape(url string) ([]models.Price, error) {
	// TODO: implement HTTP fetch and HTML parsing
	_ = url
	return nil, fmt.Errorf("not implemented")
}
