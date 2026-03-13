package models

// Price represents a product price scraped from a web page.
type Price struct {
	Name     string
	Amount   float64
	Currency string
	URL      string
}
